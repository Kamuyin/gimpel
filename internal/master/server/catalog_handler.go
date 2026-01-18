package server

import (
	"context"
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"

	gimpelv1 "gimpel/api/go/v1"
	"gimpel/internal/master/modulestore"
)

const (
	ChunkSize = 64 * 1024
)

type ModuleCatalogHandler struct {
	gimpelv1.UnimplementedModuleCatalogServiceServer

	store modulestore.Store
}

func NewModuleCatalogHandler(store modulestore.Store) *ModuleCatalogHandler {
	return &ModuleCatalogHandler{
		store: store,
	}
}

func (h *ModuleCatalogHandler) GetCatalog(ctx context.Context, req *gimpelv1.GetCatalogRequest) (*gimpelv1.GetCatalogResponse, error) {
	currentVersion := h.store.GetCatalogVersion()

	if req.CurrentVersion >= currentVersion {
		return &gimpelv1.GetCatalogResponse{
			Updated: false,
		}, nil
	}

	catalog, err := h.store.GetCatalog()
	if err != nil {
		return nil, fmt.Errorf("getting catalog: %w", err)
	}

	log.WithFields(log.Fields{
		"client_version": req.CurrentVersion,
		"server_version": currentVersion,
		"module_count":   len(catalog.Modules),
	}).Debug("serving catalog update")

	return &gimpelv1.GetCatalogResponse{
		Updated: true,
		Catalog: catalog,
	}, nil
}

func (h *ModuleCatalogHandler) GetModuleAssignments(ctx context.Context, req *gimpelv1.GetModuleAssignmentsRequest) (*gimpelv1.GetModuleAssignmentsResponse, error) {
	config, err := h.store.GetAgentAssignments(req.AgentId)
	if err != nil {
		return &gimpelv1.GetModuleAssignmentsResponse{
			Updated: false,
		}, nil
	}

	if req.CurrentVersion >= config.Version {
		return &gimpelv1.GetModuleAssignmentsResponse{
			Updated: false,
		}, nil
	}

	log.WithFields(log.Fields{
		"agent_id":       req.AgentId,
		"client_version": req.CurrentVersion,
		"server_version": config.Version,
		"assignments":    len(config.Assignments),
	}).Debug("serving module assignments")

	return &gimpelv1.GetModuleAssignmentsResponse{
		Updated: true,
		Config:  config,
	}, nil
}

func (h *ModuleCatalogHandler) DownloadModule(req *gimpelv1.DownloadModuleRequest, stream gimpelv1.ModuleCatalogService_DownloadModuleServer) error {
	module, err := h.store.GetModule(req.ModuleId, req.Version)
	if err != nil {
		return fmt.Errorf("module not found: %w", err)
	}

	reader, totalSize, err := h.store.GetModuleImage(req.ModuleId, module.Version)
	if err != nil {
		return fmt.Errorf("opening module image: %w", err)
	}
	defer reader.Close()

	log.WithFields(log.Fields{
		"module":     req.ModuleId,
		"version":    module.Version,
		"size_bytes": totalSize,
	}).Info("streaming module image")

	buf := make([]byte, ChunkSize)
	offset := int64(0)

	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading image data: %w", err)
		}

		chunk := &gimpelv1.ModuleImageChunk{
			Data:      buf[:n],
			Offset:    offset,
			TotalSize: totalSize,
			IsLast:    offset+int64(n) >= totalSize,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("sending chunk: %w", err)
		}

		offset += int64(n)
	}

	log.WithFields(log.Fields{
		"module":       req.ModuleId,
		"version":      module.Version,
		"bytes_sent":   offset,
		"chunk_count":  (offset + ChunkSize - 1) / ChunkSize,
	}).Debug("module image stream completed")

	return nil
}

func (h *ModuleCatalogHandler) VerifyModule(ctx context.Context, req *gimpelv1.VerifyModuleRequest) (*gimpelv1.VerifyModuleResponse, error) {
	module, err := h.store.GetModule(req.ModuleId, req.Version)
	if err != nil {
		return nil, fmt.Errorf("module not found: %w", err)
	}

	valid := module.Digest == req.Digest

	log.WithFields(log.Fields{
		"module":          req.ModuleId,
		"version":         req.Version,
		"client_digest":   req.Digest[:20] + "...",
		"expected_digest": module.Digest[:20] + "...",
		"valid":           valid,
	}).Debug("module verification request")

	return &gimpelv1.VerifyModuleResponse{
		Valid:     valid,
		Signature: module.Signature,
	}, nil
}
