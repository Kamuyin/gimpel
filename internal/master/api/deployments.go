package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"gimpel/internal/master/store"
)

type DeploymentAPI struct {
	store *store.Store
}

func NewDeploymentAPI(s *store.Store) *DeploymentAPI {
	return &DeploymentAPI{store: s}
}

type CreateDeploymentRequest struct {
	Modules []ModuleAssignmentRequest `json:"modules"`
}

type ModuleAssignmentRequest struct {
	ModuleID      string            `json:"module_id"`
	ModuleVersion string            `json:"module_version"`
	ExecutionMode string            `json:"execution_mode,omitempty"`
	Listeners     []ListenerRequest `json:"listeners,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
}

type ListenerRequest struct {
	ID              string `json:"id"`
	Protocol        string `json:"protocol"`
	Port            uint32 `json:"port"`
	HighInteraction bool   `json:"high_interaction,omitempty"`
}

type DeploymentResponse struct {
	SatelliteID string                    `json:"satellite_id"`
	Version     int64                     `json:"version"`
	Modules     []ModuleAssignmentInfo    `json:"modules"`
	UpdatedAt   time.Time                 `json:"updated_at"`
}

type ModuleAssignmentInfo struct {
	ModuleID      string                `json:"module_id"`
	ModuleVersion string                `json:"module_version"`
	ExecutionMode string                `json:"execution_mode"`
	Listeners     []ListenerInfo        `json:"listeners"`
	Env           map[string]string     `json:"env"`
}

type ListenerInfo struct {
	ID              string `json:"id"`
	Protocol        string `json:"protocol"`
	Port            uint32 `json:"port"`
	HighInteraction bool   `json:"high_interaction"`
}

func (da *DeploymentAPI) HandleCreateDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	satelliteID := r.PathValue("id")
	if satelliteID == "" {
		http.Error(w, "satellite id is required", http.StatusBadRequest)
		return
	}

	satellite, err := da.store.GetSatellite(satelliteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get satellite: %v", err), http.StatusInternalServerError)
		return
	}
	if satellite == nil {
		http.Error(w, "satellite not found", http.StatusNotFound)
		return
	}

	var req CreateDeploymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("failed to decode request: %v", err), http.StatusBadRequest)
		return
	}

	currentDep, err := da.store.GetDeployment(satelliteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get current deployment: %v", err), http.StatusInternalServerError)
		return
	}

	nextVersion := int64(1)
	if currentDep != nil {
		nextVersion = currentDep.Version + 1
	}

	deployment := &store.Deployment{
		SatelliteID: satelliteID,
		Version:     nextVersion,
		UpdatedAt:   time.Now(),
		Modules:     make([]store.ModuleDeployment, 0, len(req.Modules)),
	}

	for _, modReq := range req.Modules {
		listeners := make([]store.ListenerConfig, 0, len(modReq.Listeners))
		for _, l := range modReq.Listeners {
			listeners = append(listeners, store.ListenerConfig{
				ID:              l.ID,
				Protocol:        l.Protocol,
				Port:            l.Port,
				HighInteraction: l.HighInteraction,
			})
		}

		deployment.Modules = append(deployment.Modules, store.ModuleDeployment{
			ModuleID:      modReq.ModuleID,
			ModuleVersion: modReq.ModuleVersion,
			Enabled:       true,
			ExecutionMode: modReq.ExecutionMode,
			Listeners:     listeners,
			Env:           modReq.Env,
		})
	}

	if err := da.store.SetDeployment(deployment); err != nil {
		http.Error(w, fmt.Sprintf("failed to create deployment: %v", err), http.StatusInternalServerError)
		return
	}

	resp := DeploymentResponse{
		SatelliteID: deployment.SatelliteID,
		Version:     deployment.Version,
		Modules:     make([]ModuleAssignmentInfo, 0, len(deployment.Modules)),
		UpdatedAt:   deployment.UpdatedAt,
	}

	for _, mod := range deployment.Modules {
		listeners := make([]ListenerInfo, 0, len(mod.Listeners))
		for _, l := range mod.Listeners {
			listeners = append(listeners, ListenerInfo{
				ID:              l.ID,
				Protocol:        l.Protocol,
				Port:            l.Port,
				HighInteraction: l.HighInteraction,
			})
		}

		resp.Modules = append(resp.Modules, ModuleAssignmentInfo{
			ModuleID:      mod.ModuleID,
			ModuleVersion: mod.ModuleVersion,
			ExecutionMode: mod.ExecutionMode,
			Listeners:     listeners,
			Env:           mod.Env,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	log.WithFields(log.Fields{
		"satellite": satelliteID,
		"version":   nextVersion,
		"modules":   len(req.Modules),
	}).Info("deployment created")
}

func (da *DeploymentAPI) HandleGetDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	satelliteID := r.PathValue("id")
	if satelliteID == "" {
		http.Error(w, "satellite id is required", http.StatusBadRequest)
		return
	}

	deployment, err := da.store.GetDeployment(satelliteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get deployment: %v", err), http.StatusInternalServerError)
		return
	}

	if deployment == nil {
		http.Error(w, "deployment not found", http.StatusNotFound)
		return
	}

	resp := DeploymentResponse{
		SatelliteID: deployment.SatelliteID,
		Version:     deployment.Version,
		Modules:     make([]ModuleAssignmentInfo, 0, len(deployment.Modules)),
		UpdatedAt:   deployment.UpdatedAt,
	}

	for _, mod := range deployment.Modules {
		listeners := make([]ListenerInfo, 0, len(mod.Listeners))
		for _, l := range mod.Listeners {
			listeners = append(listeners, ListenerInfo{
				ID:              l.ID,
				Protocol:        l.Protocol,
				Port:            l.Port,
				HighInteraction: l.HighInteraction,
			})
		}

		resp.Modules = append(resp.Modules, ModuleAssignmentInfo{
			ModuleID:      mod.ModuleID,
			ModuleVersion: mod.ModuleVersion,
			ExecutionMode: mod.ExecutionMode,
			Listeners:     listeners,
			Env:           mod.Env,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type ListSatellitesResponse struct {
	Satellites []SatelliteInfo `json:"satellites"`
}

type SatelliteInfo struct {
	ID           string    `json:"id"`
	Hostname     string    `json:"hostname"`
	IPAddress    string    `json:"ip_address"`
	OS           string    `json:"os"`
	Arch         string    `json:"arch"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
	LastSeenAt   time.Time `json:"last_seen_at"`
}

func (da *DeploymentAPI) HandleListSatellites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	satellites, err := da.store.ListSatellites()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list satellites: %v", err), http.StatusInternalServerError)
		return
	}

	resp := ListSatellitesResponse{
		Satellites: make([]SatelliteInfo, 0, len(satellites)),
	}

	for _, sat := range satellites {
		resp.Satellites = append(resp.Satellites, SatelliteInfo{
			ID:           sat.ID,
			Hostname:     sat.Hostname,
			IPAddress:    sat.IPAddress,
			OS:           sat.OS,
			Arch:         sat.Arch,
			Status:       string(sat.Status),
			RegisteredAt: sat.RegisteredAt,
			LastSeenAt:   sat.LastSeenAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (da *DeploymentAPI) HandleGetSatellite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	satelliteID := r.PathValue("id")
	if satelliteID == "" {
		http.Error(w, "satellite id is required", http.StatusBadRequest)
		return
	}

	satellite, err := da.store.GetSatellite(satelliteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get satellite: %v", err), http.StatusInternalServerError)
		return
	}

	if satellite == nil {
		http.Error(w, "satellite not found", http.StatusNotFound)
		return
	}

	info := SatelliteInfo{
		ID:           satellite.ID,
		Hostname:     satellite.Hostname,
		IPAddress:    satellite.IPAddress,
		OS:           satellite.OS,
		Arch:         satellite.Arch,
		Status:       string(satellite.Status),
		RegisteredAt: satellite.RegisteredAt,
		LastSeenAt:   satellite.LastSeenAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (da *DeploymentAPI) HandleListDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statusFilter := r.URL.Query().Get("status")

	deployments, err := da.store.ListDeployments()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list deployments: %v", err), http.StatusInternalServerError)
		return
	}

	responses := make([]DeploymentResponse, 0, len(deployments))

	for _, dep := range deployments {
		if statusFilter != "" {
			sat, err := da.store.GetSatellite(dep.SatelliteID)
			if err != nil || sat == nil || string(sat.Status) != statusFilter {
				continue
			}
		}

		resp := DeploymentResponse{
			SatelliteID: dep.SatelliteID,
			Version:     dep.Version,
			Modules:     make([]ModuleAssignmentInfo, 0, len(dep.Modules)),
			UpdatedAt:   dep.UpdatedAt,
		}

		for _, mod := range dep.Modules {
			listeners := make([]ListenerInfo, 0, len(mod.Listeners))
			for _, l := range mod.Listeners {
				listeners = append(listeners, ListenerInfo{
					ID:              l.ID,
					Protocol:        l.Protocol,
					Port:            l.Port,
					HighInteraction: l.HighInteraction,
				})
			}

			resp.Modules = append(resp.Modules, ModuleAssignmentInfo{
				ModuleID:      mod.ModuleID,
				ModuleVersion: mod.ModuleVersion,
				ExecutionMode: mod.ExecutionMode,
				Listeners:     listeners,
				Env:           mod.Env,
			})
		}

		responses = append(responses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deployments": responses,
	})
}

func (da *DeploymentAPI) HandleDeleteDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	satelliteID := r.PathValue("id")
	if satelliteID == "" {
		http.Error(w, "satellite id is required", http.StatusBadRequest)
		return
	}

	if err := da.store.DeleteDeployment(satelliteID); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete deployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
	})

	log.WithField("satellite", satelliteID).Info("deployment deleted")
}
