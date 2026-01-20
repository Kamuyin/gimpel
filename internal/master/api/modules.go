package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/store"
	"gimpel/pkg/signing"
)

type ModuleAPI struct {
	store  *store.Store
	verifier *signing.ModuleVerifier
	keyPair  *signing.KeyPair
}

func NewModuleAPI(s *store.Store, verifier *signing.ModuleVerifier, keyPair *signing.KeyPair) *ModuleAPI {
	return &ModuleAPI{
		store:    s,
		verifier: verifier,
		keyPair:  keyPair,
	}
}

type UploadModuleRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Protocol    string `json:"protocol"`
	Labels      map[string]string `json:"labels,omitempty"`
	Signature   string `json:"signature"`
	SignedAt    int64  `json:"signed_at,omitempty"`
}

type UploadModuleResponse struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Digest    string    `json:"digest"`
	Signature string    `json:"signature,omitempty"`
	SignedBy  string    `json:"signed_by,omitempty"`
	SignedAt  int64     `json:"signed_at,omitempty"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func (ma *ModuleAPI) HandleUploadModule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(1024 * 1024 * 1024); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	id := r.FormValue("id")
	name := r.FormValue("name")
	description := r.FormValue("description")
	version := r.FormValue("version")
	protocol := r.FormValue("protocol")
	signatureHex := r.FormValue("signature")
	signedAtStr := r.FormValue("signed_at")

	if id == "" || version == "" || signatureHex == "" {
		http.Error(w, "id, version, and signature are required", http.StatusBadRequest)
		return
	}

	var signedAt int64
	if signedAtStr != "" {
		if _, err := fmt.Sscanf(signedAtStr, "%d", &signedAt); err != nil {
			http.Error(w, "invalid signed_at", http.StatusBadRequest)
			return
		}
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.WithFields(log.Fields{
		"module":   id,
		"version":  version,
		"filename": handler.Filename,
		"size":     handler.Size,
	}).Info("uploading module")

	imageMeta, err := ma.store.StoreImage(id, version, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to store image: %v", err), http.StatusInternalServerError)
		return
	}

	module := &store.Module{
		ID:          id,
		Name:        name,
		Description: description,
		Version:     version,
		Protocol:    protocol,
		Digest:      imageMeta.Digest,
		ImageRef:    fmt.Sprintf("gimpel/%s:%s", id, version),
		SizeBytes:   imageMeta.SizeBytes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if ma.verifier == nil || ma.keyPair == nil {
		http.Error(w, "module signature verification is not configured", http.StatusInternalServerError)
		return
	}

	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		http.Error(w, "invalid signature encoding", http.StatusBadRequest)
		return
	}

	moduleImage := &gimpelv1.ModuleImage{
		Id:        id,
		Version:   version,
		Digest:    imageMeta.Digest,
		Signature: signature,
		SignedBy:  ma.keyPair.KeyID,
		SignedAt:  signedAt,
	}
	if err := ma.verifier.VerifyModule(moduleImage); err != nil {
		http.Error(w, fmt.Sprintf("signature verification failed: %v", err), http.StatusBadRequest)
		return
	}

	module.Signature = signature
	module.SignedBy = ma.keyPair.KeyID
	if signedAt > 0 {
		module.SignedAt = time.Unix(signedAt, 0)
	}

	if err := ma.store.AddModule(module); err != nil {
		http.Error(w, fmt.Sprintf("failed to store module metadata: %v", err), http.StatusInternalServerError)
		return
	}

	resp := UploadModuleResponse{
		ID:        module.ID,
		Version:   module.Version,
		Digest:    module.Digest,
		Signature: fmt.Sprintf("%x", module.Signature),
		SignedBy:  module.SignedBy,
		SignedAt:  module.SignedAt.Unix(),
		Size:      module.SizeBytes,
		CreatedAt: module.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.WithFields(log.Fields{
		"module":  id,
		"version": version,
		"digest":  imageMeta.Digest,
	}).Info("module uploaded successfully")
}

type ListModulesResponse struct {
	Modules []ModuleInfo `json:"modules"`
}

type ModuleInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Protocol    string    `json:"protocol"`
	Digest      string    `json:"digest"`
	Size        int64     `json:"size_bytes"`
	SignedBy    string    `json:"signed_by,omitempty"`
	SignedAt    int64     `json:"signed_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (ma *ModuleAPI) HandleListModules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modules, err := ma.store.ListModules()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list modules: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ListModulesResponse{
		Modules: make([]ModuleInfo, 0, len(modules)),
	}

	for _, mod := range modules {
		resp.Modules = append(resp.Modules, ModuleInfo{
			ID:          mod.ID,
			Name:        mod.Name,
			Description: mod.Description,
			Version:     mod.Version,
			Protocol:    mod.Protocol,
			Digest:      mod.Digest,
			Size:        mod.SizeBytes,
			SignedBy:    mod.SignedBy,
			SignedAt:    mod.SignedAt.Unix(),
			CreatedAt:   mod.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (ma *ModuleAPI) HandleGetModule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	version := r.PathValue("version")

	if id == "" || version == "" {
		http.Error(w, "id and version are required", http.StatusBadRequest)
		return
	}

	module, err := ma.store.GetModule(id, version)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get module: %v", err), http.StatusInternalServerError)
		return
	}

	if module == nil {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}

	info := ModuleInfo{
		ID:          module.ID,
		Name:        module.Name,
		Description: module.Description,
		Version:     module.Version,
		Protocol:    module.Protocol,
		Digest:      module.Digest,
		Size:        module.SizeBytes,
		SignedBy:    module.SignedBy,
		SignedAt:    module.SignedAt.Unix(),
		CreatedAt:   module.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (ma *ModuleAPI) HandleDownloadModule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	version := r.PathValue("version")

	if id == "" || version == "" {
		http.Error(w, "id and version are required", http.StatusBadRequest)
		return
	}

	file, size, err := ma.store.OpenImage(id, version)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to open image: %v", err), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.tar.gz", id, version))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))

	if _, err := io.Copy(w, file); err != nil {
		log.WithError(err).Error("failed to write module image")
	}
}
