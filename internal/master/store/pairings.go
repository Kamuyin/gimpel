package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gimpel/pkg/storage"
)

type PairingRequest struct {
	ID             string    `json:"id"`
	Token          string    `json:"token"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	Used           bool      `json:"used"`
	UsedAt         time.Time `json:"used_at,omitempty"`
	AssignedAgent  string    `json:"assigned_agent,omitempty"`
}

func (s *Store) CreatePairingRequest(ttl time.Duration) (*PairingRequest, error) {
	id, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	token, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	pr := &PairingRequest{
		ID:        id,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
		Used:      false,
	}

	if err := s.db.PutJSON(BucketPairings, pr.ID, pr); err != nil {
		return nil, err
	}
	if err := s.db.PutJSON(BucketPairingTokens, token, map[string]string{"id": pr.ID}); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Store) GetPairingByToken(token string) (*PairingRequest, error) {
	var ref map[string]string
	if err := s.db.GetJSON(BucketPairingTokens, token, &ref); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	id := ref["id"]
	if id == "" {
		return nil, nil
	}

	var pr PairingRequest
	if err := s.db.GetJSON(BucketPairings, id, &pr); err != nil {
		if err == storage.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &pr, nil
}

func (s *Store) MarkPairingUsed(id, agentID string) error {
	var pr PairingRequest
	if err := s.db.GetJSON(BucketPairings, id, &pr); err != nil {
		return err
	}
	pr.Used = true
	pr.UsedAt = time.Now()
	pr.AssignedAgent = agentID
	return s.db.PutJSON(BucketPairings, id, &pr)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random: %w", err)
	}
	return hex.EncodeToString(b), nil
}
