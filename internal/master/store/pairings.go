package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gimpel/pkg/storage"
)

type PairingRequest struct {
	ID            string    `json:"id"`
	Token         string    `json:"token"`
	DisplayToken  string    `json:"display_token"` // Human-readable format: XXXX-XXXX
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Used          bool      `json:"used"`
	UsedAt        time.Time `json:"used_at,omitempty"`
	AssignedAgent string    `json:"assigned_agent,omitempty"`
	AgentHostname string    `json:"agent_hostname,omitempty"`
}

func generatePairingCode() (string, error) {
	const alphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := make([]byte, 8)
	for i := range code {
		code[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(code), nil
}

func formatDisplayToken(token string) string {
	if len(token) != 8 {
		return token
	}
	return token[:4] + "-" + token[4:]
}

func normalizeToken(token string) string {
	return strings.ToUpper(strings.ReplaceAll(token, "-", ""))
}

func (s *Store) CreatePairingRequest(ttl time.Duration) (*PairingRequest, error) {
	id, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	token, err := generatePairingCode()
	if err != nil {
		return nil, err
	}

	pr := &PairingRequest{
		ID:           id,
		Token:        token,
		DisplayToken: formatDisplayToken(token),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(ttl),
		Used:         false,
	}

	if err := s.db.PutJSON(BucketPairings, pr.ID, pr); err != nil {
		return nil, err
	}
	if err := s.db.PutJSON(BucketPairingTokens, normalizeToken(token), map[string]string{"id": pr.ID}); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *Store) GetPairingByToken(token string) (*PairingRequest, error) {
	normalized := normalizeToken(token)
	var ref map[string]string
	if err := s.db.GetJSON(BucketPairingTokens, normalized, &ref); err != nil {
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

func (s *Store) MarkPairingUsed(id, agentID, hostname string) error {
	var pr PairingRequest
	if err := s.db.GetJSON(BucketPairings, id, &pr); err != nil {
		return err
	}
	pr.Used = true
	pr.UsedAt = time.Now()
	pr.AssignedAgent = agentID
	pr.AgentHostname = hostname
	return s.db.PutJSON(BucketPairings, id, &pr)
}

func (s *Store) ListPairings() ([]*PairingRequest, error) {
	var pairings []*PairingRequest
	err := s.db.ForEach(BucketPairings, func(_, value []byte) error {
		var pr PairingRequest
		if err := unmarshalJSON(value, &pr); err != nil {
			return err
		}
		pairings = append(pairings, &pr)
		return nil
	})
	return pairings, err
}

func (s *Store) GetActivePairings() ([]*PairingRequest, error) {
	all, err := s.ListPairings()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var active []*PairingRequest
	for _, pr := range all {
		if !pr.Used && pr.ExpiresAt.After(now) {
			active = append(active, pr)
		}
	}
	return active, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random: %w", err)
	}
	return hex.EncodeToString(b), nil
}
