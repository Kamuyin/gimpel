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
	ID        string    `json:"id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
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
		ID:        pairing.ID,
		Token:     pairing.Token,
		ExpiresAt: pairing.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
