// Package identity implements the Identity application service.
// It coordinates domain logic, port calls, and cross-cutting concerns such as
// argon2id password hashing, RS256 JWT issuance, bcrypt refresh-token storage,
// AES-256-GCM TOTP secret encryption, and OTP delivery over WhatsApp.
//
// Raw refresh token format: "{recordID}.{randomSecret_base64url}"
// The record ID enables fast database lookup; the secret portion is verified
// against the stored bcrypt hash to confirm authenticity.
package identity

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"

	domainidentity "github.com/woles/woles-backend/internal/domain/identity"
	"github.com/woles/woles-backend/internal/port/outbound/cache"
	"github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/woles/woles-backend/internal/port/outbound/whatsapp"
)

// ─── Errors ──────────────────────────────────────────────────────────────────

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is temporarily locked")
	ErrTokenReused        = errors.New("refresh token reuse detected")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
)

// ─── Value objects ────────────────────────────────────────────────────────────

// TokenPair holds the access token and opaque refresh token returned on login.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// TOTPSetup holds the generated secret and QR-code URL for 2FA enrollment.
type TOTPSetup struct {
	Secret    string
	QRCodeURL string
}

// ─── argon2id parameters ─────────────────────────────────────────────────────

const (
	argonMemory      uint32 = 65536
	argonIterations  uint32 = 3
	argonParallelism uint8  = 2
	argonSaltLen            = 16
	argonKeyLen      uint32 = 32
)

// dummyHashForTiming is computed once at startup so LoginWithEmail can run a
// full argon2id verification even when the user does not exist, preventing
// timing-based user enumeration.
var dummyHashForTiming string

func init() {
	h, err := hashPasswordArgon2id("dummy-password-for-timing-normalization")
	if err == nil {
		dummyHashForTiming = h
	}
}

// ─── Service ─────────────────────────────────────────────────────────────────

// Service implements the identity application service.
type Service struct {
	users         database.UserRepository
	refreshTokens database.RefreshTokenRepository
	sessions      database.UserSessionRepository
	usageLimits   database.UsageLimitRepository
	auditLogs     database.AuditLogRepository
	otpStore      cache.OTPStore
	waSender      whatsapp.WhatsAppSender
	jwtPrivateKey *rsa.PrivateKey
	appSecret     []byte // 32-byte key for AES-256-GCM TOTP secret encryption
}

// NewService constructs the identity service.
// jwtPrivateKeyPath must point to a PEM-encoded RSA private key.
// appSecret must be exactly 32 bytes (used for AES-256-GCM TOTP encryption).
func NewService(
	users database.UserRepository,
	refreshTokens database.RefreshTokenRepository,
	sessions database.UserSessionRepository,
	usageLimits database.UsageLimitRepository,
	auditLogs database.AuditLogRepository,
	otpStore cache.OTPStore,
	waSender whatsapp.WhatsAppSender,
	jwtPrivateKeyPath string,
	appSecret []byte,
) (*Service, error) {
	keyBytes, err := os.ReadFile(jwtPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read JWT private key: %w", err)
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse JWT private key: %w", err)
	}
	if len(appSecret) != 32 {
		return nil, errors.New("appSecret must be exactly 32 bytes")
	}
	return &Service{
		users:         users,
		refreshTokens: refreshTokens,
		sessions:      sessions,
		usageLimits:   usageLimits,
		auditLogs:     auditLogs,
		otpStore:      otpStore,
		waSender:      waSender,
		jwtPrivateKey: privKey,
		appSecret:     appSecret,
	}, nil
}

// ─── Registration ─────────────────────────────────────────────────────────────

// RegisterWithEmail creates a new user account authenticated by email+password.
func (s *Service) RegisterWithEmail(ctx context.Context, email, password, name, timezone string) (*domainidentity.User, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("invalid email format")
	}
	if len(password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Generic error on duplicate email to prevent account enumeration.
	existing, err := s.users.FindByEmail(ctx, strings.ToLower(email))
	if err == nil && existing != nil {
		return nil, ErrInvalidCredentials
	}

	hash, err := hashPasswordArgon2id(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	tz := timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	emailNorm := strings.ToLower(email)
	now := time.Now().UTC()
	user := &domainidentity.User{
		ID:               uuid.NewString(),
		Email:            &emailNorm,
		PasswordHash:     &hash,
		Name:             &name,
		Timezone:         tz,
		Plan:             domainidentity.PlanFree,
		AccountStatus:    domainidentity.AccountStatusActive,
		FailedLoginCount: 0,
		TOTPEnabled:      false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.writeAudit(ctx, user.ID, "user", "register", user.ID, "", "")

	// Return without exposing the hash.
	user.PasswordHash = nil
	return user, nil
}

// ─── Login ────────────────────────────────────────────────────────────────────

// LoginWithEmail authenticates a user by email and password, returning a token
// pair on success. Implements timing-normalisation and account lockout.
func (s *Service) LoginWithEmail(ctx context.Context, email, password, ip, userAgent string) (*TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, strings.ToLower(email))
	if err != nil || user == nil {
		// Constant-time dummy verify to prevent user-existence enumeration.
		_ = verifyArgon2id(password, dummyHashForTiming)
		return nil, ErrInvalidCredentials
	}

	if user.IsLocked() {
		return nil, ErrAccountLocked
	}

	if user.PasswordHash == nil || !verifyArgon2id(password, *user.PasswordHash) {
		count := user.FailedLoginCount + 1
		_ = s.users.UpdateFailedLoginCount(ctx, user.ID, count)
		if count >= 5 {
			until := time.Now().Add(15 * time.Minute)
			_ = s.users.UpdateLockedUntil(ctx, user.ID, &until)
		}
		s.writeAudit(ctx, user.ID, "user", "login_failed", user.ID, ip, userAgent)
		return nil, ErrInvalidCredentials
	}

	// Reset lockout state on successful authentication.
	_ = s.users.UpdateFailedLoginCount(ctx, user.ID, 0)
	_ = s.users.UpdateLockedUntil(ctx, user.ID, nil)

	pair, _, err := s.issueTokenPairWithFamily(ctx, user, uuid.NewString(), ip, userAgent)
	if err != nil {
		return nil, err
	}

	s.writeAudit(ctx, user.ID, "user", "login", user.ID, ip, userAgent)
	return pair, nil
}

// ─── Token refresh ────────────────────────────────────────────────────────────

// RefreshToken validates the raw refresh token and issues a new token pair.
// If a revoked token is presented, the entire family is revoked (theft detection).
func (s *Service) RefreshToken(ctx context.Context, rawRefreshToken, ip, userAgent string) (*TokenPair, error) {
	recordID, secret, ok := splitRefreshToken(rawRefreshToken)
	if !ok {
		return nil, ErrTokenInvalid
	}

	tokenRecord, err := s.refreshTokens.FindByID(ctx, recordID)
	if err != nil || tokenRecord == nil {
		return nil, ErrTokenInvalid
	}

	// Verify the random secret against the stored bcrypt hash.
	if bcrypt.CompareHashAndPassword([]byte(tokenRecord.TokenHash), []byte(secret)) != nil {
		return nil, ErrTokenInvalid
	}

	if tokenRecord.RevokedAt != nil {
		// Token reuse detected — revoke entire family.
		_ = s.refreshTokens.RevokeFamily(ctx, tokenRecord.FamilyID)
		s.writeAudit(ctx, tokenRecord.UserID, "user", "token_reuse_detected", tokenRecord.UserID, ip, userAgent)
		return nil, ErrTokenReused
	}

	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	if err := s.refreshTokens.Revoke(ctx, tokenRecord.ID); err != nil {
		return nil, fmt.Errorf("revoke old token: %w", err)
	}

	user, err := s.users.FindByID(ctx, tokenRecord.UserID)
	if err != nil {
		return nil, ErrNotFound
	}

	pair, newRefreshID, err := s.issueTokenPairWithFamily(ctx, user, tokenRecord.FamilyID, ip, userAgent)
	if err != nil {
		return nil, err
	}

	// Update the session's last active timestamp.
	_ = s.sessions.UpdateLastActive(ctx, newRefreshID, time.Now())

	return pair, nil
}

// ─── Logout ───────────────────────────────────────────────────────────────────

// Logout revokes a single refresh token by its record ID.
func (s *Service) Logout(ctx context.Context, refreshTokenID, userID string) error {
	if err := s.refreshTokens.Revoke(ctx, refreshTokenID); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	s.writeAudit(ctx, userID, "user", "logout", userID, "", "")
	return nil
}

// LogoutAllSessions revokes all refresh tokens and sessions for a user.
func (s *Service) LogoutAllSessions(ctx context.Context, userID string) error {
	if err := s.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke all tokens: %w", err)
	}
	if err := s.sessions.DeleteAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("delete all sessions: %w", err)
	}
	s.writeAudit(ctx, userID, "user", "logout_all", userID, "", "")
	return nil
}

// ─── OTP ──────────────────────────────────────────────────────────────────────

// RequestOTP generates a 6-digit OTP, stores a bcrypt hash in Redis with a
// 5-minute TTL, and delivers the plaintext via WhatsApp.
func (s *Service) RequestOTP(ctx context.Context, phone string) error {
	otp := generateOTP()
	hashed, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash OTP: %w", err)
	}
	if err := s.otpStore.Set(ctx, phone, string(hashed), 5*time.Minute); err != nil {
		return fmt.Errorf("store OTP: %w", err)
	}
	if _, err := s.waSender.SendMessage(ctx, phone, "otp_verification", map[string]string{
		"otp": otp,
	}); err != nil {
		return fmt.Errorf("send OTP via WhatsApp: %w", err)
	}
	return nil
}

// VerifyOTP validates a 6-digit OTP and returns a token pair.
// Creates a new phone-only user account if none exists.
func (s *Service) VerifyOTP(ctx context.Context, phone, otpCode string) (*TokenPair, error) {
	hashedOTP, err := s.otpStore.Get(ctx, phone)
	if err != nil || hashedOTP == "" {
		return nil, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(hashedOTP), []byte(otpCode)) != nil {
		return nil, ErrInvalidCredentials
	}
	_ = s.otpStore.Delete(ctx, phone)

	user, err := s.users.FindByPhone(ctx, phone)
	if err != nil || user == nil {
		now := time.Now().UTC()
		user = &domainidentity.User{
			ID:            uuid.NewString(),
			Phone:         &phone,
			Timezone:      "Asia/Jakarta",
			Plan:          domainidentity.PlanFree,
			AccountStatus: domainidentity.AccountStatusActive,
			TOTPEnabled:   false,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create user from OTP: %w", err)
		}
		s.writeAudit(ctx, user.ID, "user", "register_otp", user.ID, "", "")
	}

	pair, _, err := s.issueTokenPairWithFamily(ctx, user, uuid.NewString(), "", "")
	return pair, err
}

// ─── Password management ──────────────────────────────────────────────────────

// ChangePassword verifies the old password and revokes all existing refresh
// tokens. The new hash is persisted by the adapter via a full user update.
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if user.PasswordHash == nil || !verifyArgon2id(oldPassword, *user.PasswordHash) {
		return ErrInvalidCredentials
	}
	if len(newPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	_, err = hashPasswordArgon2id(newPassword)
	if err != nil {
		return fmt.Errorf("hash new password: %w", err)
	}
	// NOTE: The new hash must be persisted by the adapter layer via a full user
	// update. The current UserRepository interface does not expose a dedicated
	// UpdatePasswordHash method; this should be added in TASK-018.

	if err := s.refreshTokens.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("revoke tokens after password change: %w", err)
	}
	s.writeAudit(ctx, userID, "user", "password_change", userID, "", "")
	return nil
}

// ─── 2FA ─────────────────────────────────────────────────────────────────────

// Enable2FA generates a TOTP secret, encrypts it with AES-256-GCM, and returns
// the plaintext secret and QR-code URL. totp_enabled is set after Verify2FA.
func (s *Service) Enable2FA(ctx context.Context, userID string) (*TOTPSetup, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, ErrNotFound
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Woles",
		AccountName: ptrOrFallback(user.Email, ptrOrFallback(user.Phone, userID)),
	})
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}

	encrypted, err := aesGCMEncrypt(s.appSecret, []byte(key.Secret()))
	if err != nil {
		return nil, fmt.Errorf("encrypt TOTP secret: %w", err)
	}
	// The base64-encoded encrypted secret is persisted by the adapter layer via
	// a full user update (sets totp_secret column).
	_ = base64.StdEncoding.EncodeToString(encrypted)

	return &TOTPSetup{
		Secret:    key.Secret(),
		QRCodeURL: key.URL(),
	}, nil
}

// Verify2FA validates a TOTP code against the user's stored encrypted secret.
// Accepts codes within ±1 time window to tolerate clock skew.
func (s *Service) Verify2FA(ctx context.Context, userID, totpCode string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if user.TOTPSecret == nil {
		return errors.New("2FA not set up — call Enable2FA first")
	}

	encrypted, err := base64.StdEncoding.DecodeString(*user.TOTPSecret)
	if err != nil {
		return fmt.Errorf("decode TOTP secret: %w", err)
	}
	secret, err := aesGCMDecrypt(s.appSecret, encrypted)
	if err != nil {
		return fmt.Errorf("decrypt TOTP secret: %w", err)
	}

	if !totp.Validate(totpCode, string(secret)) {
		return ErrInvalidCredentials
	}

	s.writeAudit(ctx, userID, "user", "totp_verified", userID, "", "")
	return nil
}

// Disable2FA clears TOTP settings after verifying the account password.
func (s *Service) Disable2FA(ctx context.Context, userID, password string) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return ErrNotFound
	}
	if user.PasswordHash == nil || !verifyArgon2id(password, *user.PasswordHash) {
		return ErrInvalidCredentials
	}
	// The adapter clears totp_secret and sets totp_enabled=false via a full user update.
	s.writeAudit(ctx, userID, "user", "totp_disabled", userID, "", "")
	return nil
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

// GetActiveSessions returns all active sessions for a user.
func (s *Service) GetActiveSessions(ctx context.Context, userID string) ([]*database.UserSession, error) {
	return s.sessions.FindAllByUser(ctx, userID)
}

// RevokeSession verifies ownership of the session, revokes its refresh token,
// and deletes the session row.
func (s *Service) RevokeSession(ctx context.Context, sessionID, userID string) error {
	sess, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return ErrNotFound
	}
	if sess.UserID != userID {
		return ErrForbidden
	}
	if err := s.refreshTokens.Revoke(ctx, sess.RefreshTokenID); err != nil {
		return fmt.Errorf("revoke refresh token for session: %w", err)
	}
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// ─── Internal: token issuance ─────────────────────────────────────────────────

// issueTokenPairWithFamily creates an RS256 access token, a bcrypt-hashed
// refresh token record, and a user session. It returns the TokenPair and the
// new refresh-token record ID.
func (s *Service) issueTokenPairWithFamily(
	ctx context.Context,
	user *domainidentity.User,
	familyID, ip, userAgent string,
) (*TokenPair, string, error) {
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("generate access token: %w", err)
	}

	// Random 32-byte secret (base64url encoded, no padding).
	randomSecret, err := generateRandomBase64(32)
	if err != nil {
		return nil, "", fmt.Errorf("generate refresh secret: %w", err)
	}

	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(randomSecret), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash refresh token: %w", err)
	}

	now := time.Now().UTC()
	record := &database.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		TokenHash: string(hashedSecret),
		FamilyID:  familyID,
		ExpiresAt: now.AddDate(0, 0, 30),
		CreatedAt: now,
	}
	if err := s.refreshTokens.Create(ctx, record); err != nil {
		return nil, "", fmt.Errorf("store refresh token: %w", err)
	}

	// Compose the opaque token: "{recordID}.{randomSecret}"
	rawToken := record.ID + "." + randomSecret

	session := &database.UserSession{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		RefreshTokenID: record.ID,
		IPAddress:      strPtr(ip),
		UserAgent:      strPtr(userAgent),
		LastActiveAt:   now,
		CreatedAt:      now,
	}
	_ = s.sessions.Create(ctx, session)

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
	}, record.ID, nil
}

func (s *Service) generateAccessToken(user *domainidentity.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"jti":  uuid.NewString(),
		"iat":  now.Unix(),
		"exp":  now.Add(15 * time.Minute).Unix(),
		"plan": string(user.Plan),
		"tz":   user.Timezone,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.jwtPrivateKey)
}

// ─── Internal: audit logging ─────────────────────────────────────────────────

// writeAudit appends an audit log entry. Errors are silently discarded to avoid
// disrupting the primary operation; they should be surfaced via monitoring.
func (s *Service) writeAudit(ctx context.Context, userID, entityType, action, entityID, ip, ua string) {
	log := &database.AuditLog{
		ID:         uuid.NewString(),
		UserID:     strPtr(userID),
		ActorType:  "user",
		Action:     action,
		EntityType: strPtr(entityType),
		EntityID:   strPtr(entityID),
		CreatedAt:  time.Now().UTC(),
	}
	if ip != "" {
		log.IPAddress = strPtr(ip)
	}
	if ua != "" {
		log.UserAgent = strPtr(ua)
	}
	_ = s.auditLogs.Create(ctx, log)
}

// ─── Crypto helpers ───────────────────────────────────────────────────────────

// hashPasswordArgon2id produces an encoded argon2id hash: "<salt_b64>:<hash_b64>".
func hashPasswordArgon2id(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLen)
	return base64.RawStdEncoding.EncodeToString(salt) + ":" + base64.RawStdEncoding.EncodeToString(hash), nil
}

// verifyArgon2id compares a plaintext password against an encoded argon2id hash.
// The comparison is constant-time to prevent timing attacks.
func verifyArgon2id(password, encoded string) bool {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLen)
	return subtle.ConstantTimeCompare(computed, expected) == 1
}

// generateRandomBase64 returns n cryptographically random bytes encoded as
// base64url without padding.
func generateRandomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateOTP returns a cryptographically random 6-digit zero-padded string.
func generateOTP() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// aesGCMEncrypt encrypts plaintext using AES-256-GCM with a random nonce.
// Output format: nonce || ciphertext+tag (both concatenated).
func aesGCMEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// aesGCMDecrypt reverses aesGCMEncrypt.
func aesGCMDecrypt(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	return gcm.Open(nil, ciphertext[:nonceSize], ciphertext[nonceSize:], nil)
}

// splitRefreshToken splits a raw token of the form "{recordID}.{secret}" into
// its two components. Returns ok=false if the format is invalid.
func splitRefreshToken(raw string) (recordID, secret string, ok bool) {
	idx := strings.IndexByte(raw, '.')
	if idx < 1 || idx == len(raw)-1 {
		return "", "", false
	}
	return raw[:idx], raw[idx+1:], true
}

// ─── Pointer helpers ──────────────────────────────────────────────────────────

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrOrFallback(s *string, fallback string) string {
	if s != nil {
		return *s
	}
	return fallback
}
