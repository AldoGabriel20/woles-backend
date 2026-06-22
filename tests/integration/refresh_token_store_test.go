package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	portdb "github.com/woles/woles-backend/internal/port/outbound/database"
)

func TestRefreshTokenStore_IssueRotateRevokeReuseDetection(t *testing.T) {
	db := skipIfNoDB(t)
	repo := postgres.NewRefreshTokenRepo(db.Pool)
	ctx := context.Background()

	userID := uuid.NewString()
	familyID := uuid.NewString()

	// Issue token.
	tok := &portdb.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: "hash-v1",
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := repo.Create(ctx, tok); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Fetch and verify.
	fetched, err := repo.FindByID(ctx, tok.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if fetched.TokenHash != tok.TokenHash {
		t.Errorf("TokenHash mismatch")
	}

	// Rotate: revoke old token, create new one.
	if err := repo.Revoke(ctx, tok.ID); err != nil {
		t.Fatalf("Revoke: %v", err)
	}
	newTok := &portdb.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: "hash-v2",
		FamilyID:  familyID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := repo.Create(ctx, newTok); err != nil {
		t.Fatalf("Create new token: %v", err)
	}

	// Old token should be revoked.
	old, _ := repo.FindByID(ctx, tok.ID)
	if old.RevokedAt == nil {
		t.Error("old token should have RevokedAt set after revocation")
	}

	// Reuse detection: revoke entire family.
	if err := repo.RevokeFamily(ctx, familyID); err != nil {
		t.Fatalf("RevokeFamily: %v", err)
	}
	newFetched, _ := repo.FindByID(ctx, newTok.ID)
	if newFetched.RevokedAt == nil {
		t.Error("new token in family should also be revoked after family revocation")
	}
}
