package unit_test

import (
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ─── Helpers that mirror the identity service token rotation logic ────────────

// hashRefreshToken bcrypt-hashes the raw secret portion of a refresh token.
func hashRefreshToken(raw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(hash), err
}

func generateRefreshSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// ─── Token record (in-memory stand-in) ───────────────────────────────────────

type tokenRecord struct {
	ID        string
	Hash      string
	FamilyID  string
	Revoked   bool
	ExpiresAt time.Time
}

type tokenStore struct {
	records map[string]*tokenRecord
}

func newTokenStore() *tokenStore { return &tokenStore{records: map[string]*tokenRecord{}} }

func (s *tokenStore) issue(familyID string) (raw string, rec *tokenRecord) {
	raw = generateRefreshSecret()
	hash, _ := hashRefreshToken(raw)
	rec = &tokenRecord{
		ID:        generateRefreshSecret()[:8],
		Hash:      hash,
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	s.records[rec.ID] = rec
	return raw, rec
}

func (s *tokenStore) findByID(id string) *tokenRecord { return s.records[id] }

func (s *tokenStore) revoke(id string) { s.records[id].Revoked = true }

func (s *tokenStore) revokeFamily(familyID string) {
	for _, r := range s.records {
		if r.FamilyID == familyID {
			r.Revoked = true
		}
	}
}

// rotate issues a new token and revokes the old one, simulating the service.
// Returns error if old token was already revoked (reuse detection).
func (s *tokenStore) rotate(oldID, oldRaw string, familyID string) (string, *tokenRecord, error) {
	old := s.findByID(oldID)
	if old == nil {
		return "", nil, errTokenNotFound
	}
	if old.Revoked {
		// Reuse detected: revoke entire family.
		s.revokeFamily(familyID)
		return "", nil, errTokenReusedLocal
	}
	if bcrypt.CompareHashAndPassword([]byte(old.Hash), []byte(oldRaw)) != nil {
		return "", nil, errTokenInvalidLocal
	}
	s.revoke(oldID)
	newRaw, newRec := s.issue(familyID)
	return newRaw, newRec, nil
}

var (
	errTokenNotFound     = errStr("token not found")
	errTokenReusedLocal  = errStr("refresh token reuse detected")
	errTokenInvalidLocal = errStr("refresh token invalid")
)

type errStr string

func (e errStr) Error() string { return string(e) }

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestRefreshTokenRotation_NormalFlow(t *testing.T) {
	store := newTokenStore()
	familyID := "family-1"
	raw, rec := store.issue(familyID)

	newRaw, _, err := store.rotate(rec.ID, raw, familyID)
	if err != nil {
		t.Fatalf("normal rotation failed: %v", err)
	}
	if newRaw == raw {
		t.Error("new token should differ from old token")
	}
}

func TestRefreshTokenRotation_OldTokenRejectedAfterRotation(t *testing.T) {
	store := newTokenStore()
	familyID := "family-2"
	raw, rec := store.issue(familyID)
	_, newRec, _ := store.rotate(rec.ID, raw, familyID)

	// Try to use old token again.
	_, _, err := store.rotate(rec.ID, raw, familyID)
	if err == nil {
		t.Error("old token should be rejected after rotation")
	}
	// New token should still be valid.
	old := store.findByID(rec.ID)
	if !old.Revoked {
		t.Error("old record should be revoked")
	}
	_ = newRec
}

func TestRefreshTokenRotation_ReuseRevokesFamily(t *testing.T) {
	store := newTokenStore()
	familyID := "family-3"
	raw1, rec1 := store.issue(familyID)
	raw2, rec2, _ := store.rotate(rec1.ID, raw1, familyID)

	// Attempt to reuse the original (already revoked) token.
	_, _, err := store.rotate(rec1.ID, raw1, familyID)
	if err == nil {
		t.Error("reuse should be detected and return error")
	}
	// Entire family should now be revoked.
	if !store.findByID(rec2.ID).Revoked {
		t.Error("second token in family should be revoked after reuse detection")
	}
	_ = raw2
}
