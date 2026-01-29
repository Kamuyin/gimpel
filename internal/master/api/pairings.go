package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gimpel/internal/master/store"
)

type PairingAPI struct {
	store *store.Store
}

type CreatePairingRequest struct {
	TTLSeconds int64 `json:"ttl_seconds"`
}

type PairingResponse struct {
	ID           string    `json:"id"`
	Token        string    `json:"token"`         // Raw token for API use
	DisplayToken string    `json:"display_token"` // Human-readable: XXXX-XXXX
	ExpiresAt    time.Time `json:"expires_at"`
	ExpiresIn    int       `json:"expires_in_seconds"`
}

type PairingInfo struct {
	ID            string    `json:"id"`
	DisplayToken  string    `json:"display_token"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Used          bool      `json:"used"`
	UsedAt        time.Time `json:"used_at,omitempty"`
	AssignedAgent string    `json:"assigned_agent,omitempty"`
	AgentHostname string    `json:"agent_hostname,omitempty"`
}

func NewPairingAPI(s *store.Store) *PairingAPI {
	return &PairingAPI{store: s}
}

func (pa *PairingAPI) HandleCreatePairing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePairingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	ttl := 10 * time.Minute
	if req.TTLSeconds > 0 {
		ttl = time.Duration(req.TTLSeconds) * time.Second
	}

	pairing, err := pa.store.CreatePairingRequest(ttl)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create pairing: %v", err), http.StatusInternalServerError)
		return
	}

	resp := PairingResponse{
		ID:           pairing.ID,
		Token:        pairing.Token,
		DisplayToken: pairing.DisplayToken,
		ExpiresAt:    pairing.ExpiresAt,
		ExpiresIn:    int(time.Until(pairing.ExpiresAt).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (pa *PairingAPI) HandleListPairings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pairings, err := pa.store.ListPairings()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to list pairings: %v", err), http.StatusInternalServerError)
		return
	}

	infos := make([]PairingInfo, 0, len(pairings))
	for _, p := range pairings {
		infos = append(infos, PairingInfo{
			ID:            p.ID,
			DisplayToken:  p.DisplayToken,
			CreatedAt:     p.CreatedAt,
			ExpiresAt:     p.ExpiresAt,
			Used:          p.Used,
			UsedAt:        p.UsedAt,
			AssignedAgent: p.AssignedAgent,
			AgentHostname: p.AgentHostname,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pairings": infos})
}

func (pa *PairingAPI) HandleGetActivePairings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pairings, err := pa.store.GetActivePairings()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get active pairings: %v", err), http.StatusInternalServerError)
		return
	}

	infos := make([]PairingInfo, 0, len(pairings))
	for _, p := range pairings {
		infos = append(infos, PairingInfo{
			ID:           p.ID,
			DisplayToken: p.DisplayToken,
			CreatedAt:    p.CreatedAt,
			ExpiresAt:    p.ExpiresAt,
			Used:         p.Used,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"pairings": infos})
}
