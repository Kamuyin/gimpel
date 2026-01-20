package server

import (
	"context"
	"fmt"
	"io"
	"time"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/store"
)

const (
	ChunkSize = 64 * 1024
)

type ModuleCatalogHandler struct {
	gimpelv1.UnimplementedModuleCatalogServiceServer

	store  *store.Store
}

func NewModuleCatalogHandler(s *store.Store) *ModuleCatalogHandler {
	return &ModuleCatalogHandler{
		store:  s,
	}
}

func (h *ModuleCatalogHandler) GetCatalog(ctx context.Context, req *gimpelv1.GetCatalogRequest) (*gimpelv1.GetCatalogResponse, error) {
	modules, err := h.store.ListModules()
	if err != nil {
		return nil, fmt.Errorf("listing modules: %w", err)
	}

	catalog := &gimpelv1.ModuleCatalog{
		Version:   time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
		Modules:   make([]*gimpelv1.ModuleImage, 0, len(modules)),
	}

	for _, mod := range modules {
		catalog.Modules = append(catalog.Modules, &gimpelv1.ModuleImage{
			Id:        mod.ID,
			Version:   mod.Version,
			Digest:    mod.Digest,
			Signature: mod.Signature,
			SignedBy:  mod.SignedBy,
			SignedAt:  mod.SignedAt.Unix(),
			SizeBytes: mod.SizeBytes,
		})
	}

	log.WithFields(log.Fields{
		"client_version": req.CurrentVersion,
		"server_version": catalog.Version,
		"module_count":   len(catalog.Modules),
	}).Debug("serving catalog")

	return &gimpelv1.GetCatalogResponse{
		Updated: true,
		Catalog: catalog,
	}, nil
}

func (h *ModuleCatalogHandler) GetModuleAssignments(ctx context.Context, req *gimpelv1.GetModuleAssignmentsRequest) (*gimpelv1.GetModuleAssignmentsResponse, error) {
	deployment, err := h.store.GetDeployment(req.AgentId)
	if err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	if deployment == nil {
		return &gimpelv1.GetModuleAssignmentsResponse{
			Updated: false,
		}, nil
	}

	if req.CurrentVersion >= deployment.Version {
		return &gimpelv1.GetModuleAssignmentsResponse{
			Updated: false,
		}, nil
	}

	config := &gimpelv1.AgentModuleConfig{
		AgentId:     req.AgentId,
		Version:     deployment.Version,
		Assignments: make([]*gimpelv1.ModuleAssignment, 0, len(deployment.Modules)),
	}

	for _, mod := range deployment.Modules {
		listeners := make([]*gimpelv1.ListenerAssignment, 0, len(mod.Listeners))
		for _, l := range mod.Listeners {
			listeners = append(listeners, &gimpelv1.ListenerAssignment{
				Id:              l.ID,
				Protocol:        l.Protocol,
				Port:            l.Port,
				HighInteraction: l.HighInteraction,
			})
		}

		config.Assignments = append(config.Assignments, &gimpelv1.ModuleAssignment{
			ModuleId:      mod.ModuleID,
			Version:       mod.ModuleVersion,
			ExecutionMode: mod.ExecutionMode,
			Listeners:     listeners,
			Env:           mod.Env,
		})
	}

	log.WithFields(log.Fields{
		"agent_id":    req.AgentId,
		"version":     deployment.Version,
		"assignments": len(config.Assignments),
	}).Debug("serving assignments")

	return &gimpelv1.GetModuleAssignmentsResponse{
		Updated: true,
		Config:  config,
	}, nil
}

func (h *ModuleCatalogHandler) DownloadModule(req *gimpelv1.DownloadModuleRequest, stream gimpelv1.ModuleCatalogService_DownloadModuleServer) error {
	reader, size, err := h.store.OpenImage(req.ModuleId, req.Version)
	if err != nil {
		return fmt.Errorf("opening image: %w", err)
	}
	defer reader.Close()

	log.WithFields(log.Fields{
		"module":     req.ModuleId,
		"version":    req.Version,
		"size_bytes": size,
	}).Info("streaming module image")

	buf := make([]byte, ChunkSize)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading image: %w", err)
		}

		if err := stream.Send(&gimpelv1.ModuleImageChunk{
			Data: buf[:n],
		}); err != nil {
			return fmt.Errorf("sending chunk: %w", err)
		}
	}

	return nil
}

func (h *ModuleCatalogHandler) VerifyModule(ctx context.Context, req *gimpelv1.VerifyModuleRequest) (*gimpelv1.VerifyModuleResponse, error) {
	meta, err := h.store.GetImage(req.ModuleId, req.Version)
	if err != nil {
		return nil, fmt.Errorf("getting image: %w", err)
	}

	if meta == nil {
		return &gimpelv1.VerifyModuleResponse{
			Valid: false,
		}, nil
	}

	if meta.Digest != req.Digest {
		return &gimpelv1.VerifyModuleResponse{
			Valid: false,
		}, nil
	}

	mod, _ := h.store.GetModule(req.ModuleId, req.Version)
	var sig []byte
	if mod != nil {
		sig = mod.Signature
	}

	resp := &gimpelv1.VerifyModuleResponse{
		Valid:     true,
		Signature: sig,
	}
	
	if mod != nil {
		resp.SignedBy = mod.SignedBy
		resp.SignedAt = mod.SignedAt.Unix()
	}
	
	return resp, nil
}
