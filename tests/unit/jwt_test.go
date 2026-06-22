package unit_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return key
}

func signRS256Token(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func parseToken(t *testing.T, tokenStr string, pub *rsa.PublicKey) (*jwt.Token, error) {
	t.Helper()
	return jwt.Parse(tokenStr, func(tok *jwt.Token) (interface{}, error) {
		if _, ok := tok.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return pub, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithExpirationRequired())
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestJWT_ValidTokenParsesSuccessfully(t *testing.T) {
	key := generateTestRSAKey(t)
	claims := jwt.MapClaims{
		"sub":  "user-123",
		"plan": "free",
		"tz":   "Asia/Jakarta",
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}
	tokenStr := signRS256Token(t, key, claims)
	tok, err := parseToken(t, tokenStr, &key.PublicKey)
	if err != nil || !tok.Valid {
		t.Errorf("valid token should parse successfully: %v", err)
	}
}

func TestJWT_ExpiredTokenRejected(t *testing.T) {
	key := generateTestRSAKey(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-1 * time.Minute).Unix(),
	}
	tokenStr := signRS256Token(t, key, claims)
	_, err := parseToken(t, tokenStr, &key.PublicKey)
	if err == nil {
		t.Error("expired token should be rejected")
	}
}

func TestJWT_TamperedPayloadRejected(t *testing.T) {
	key := generateTestRSAKey(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	tokenStr := signRS256Token(t, key, claims)
	// Corrupt the payload segment (index 1)
	parts := splitToken(tokenStr)
	if len(parts) != 3 {
		t.Fatal("expected 3 JWT parts")
	}
	parts[1] = "dGFtcGVyZWQ" // base64url("tampered")
	tampered := parts[0] + "." + parts[1] + "." + parts[2]
	_, err := parseToken(t, tampered, &key.PublicKey)
	if err == nil {
		t.Error("tampered payload should be rejected")
	}
}

func TestJWT_NoneAlgorithmRejected(t *testing.T) {
	// Craft a token with alg=none manually.
	noneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEyMyIsImV4cCI6OTk5OTk5OTk5OX0."
	key := generateTestRSAKey(t)
	_, err := parseToken(t, noneToken, &key.PublicKey)
	if err == nil {
		t.Error("none algorithm token should be rejected")
	}
}

func TestJWT_WrongKeyRejected(t *testing.T) {
	key1 := generateTestRSAKey(t)
	key2 := generateTestRSAKey(t)
	claims := jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	}
	tokenStr := signRS256Token(t, key1, claims)
	// Verify with a different public key.
	_, err := parseToken(t, tokenStr, &key2.PublicKey)
	if err == nil {
		t.Error("token signed with key1 should be rejected by key2")
	}
}

// splitToken splits a JWT string into its three dot-separated parts.
func splitToken(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
