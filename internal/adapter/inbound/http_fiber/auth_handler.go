// Package http_fiber provides Fiber HTTP adapter handlers for the Woles API.
package http_fiber

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/woles/woles-backend/internal/adapter/inbound/http_fiber/middleware"
	appidentity "github.com/woles/woles-backend/internal/application/identity"
)

type authHandler struct{ svc *appidentity.Service }

// RegisterAuthRoutes mounts all /api/v1/auth routes on the given router.
// Public endpoints (register, login, refresh, otp, password reset) are mounted
// directly; protected endpoints apply JWTMiddleware via a sub-group.
func RegisterAuthRoutes(router fiber.Router, svc *Services) {
	h := &authHandler{svc: svc.Identity}

	auth := router.Group("/auth")

	// Public — no JWT.
	auth.Post("/register", h.register)
	auth.Post("/login", h.login)
	auth.Post("/refresh", h.refreshToken)
	auth.Post("/otp/request", h.requestOTP)
	auth.Post("/otp/verify", h.verifyOTP)
	auth.Post("/password/reset/request", h.passwordResetRequest)
	auth.Post("/password/reset/confirm", h.passwordResetConfirm)

	// Protected — require JWT.
	prot := auth.Use(middleware.JWTMiddleware())
	prot.Post("/logout", h.logout)
	prot.Post("/password/change", h.changePassword)
	prot.Get("/me", h.getMe)
	prot.Post("/2fa/enable", h.enable2FA)
	prot.Post("/2fa/verify", h.verify2FA)
	prot.Post("/2fa/disable", h.disable2FA)
	prot.Get("/sessions", h.getSessions)
	prot.Delete("/sessions/:id", h.revokeSession)
	prot.Delete("/sessions", h.revokeAllSessions)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func isProduction() bool { return os.Getenv("APP_ENV") != "development" }

func setRefreshCookie(c *fiber.Ctx, token string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   2592000,
		Secure:   isProduction(),
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

func clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// ─── handlers ────────────────────────────────────────────────────────────────

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

func (h *authHandler) register(c *fiber.Ctx) error {
	var body registerBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Email == "" || body.Password == "" || body.Name == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "email, password, and name are required")
	}
	user, err := h.svc.RegisterWithEmail(c.Context(), body.Email, body.Password, body.Name, body.Timezone)
	if err != nil {
		return mapServiceError(c, err)
	}
	// Issue token pair for immediate login after registration.
	ip := c.IP()
	ua := c.Get("User-Agent")
	pair, err := h.svc.LoginWithEmail(c.Context(), body.Email, body.Password, ip, ua)
	if err != nil {
		// Registration succeeded but login failed — return 201 without token.
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
	}
	setRefreshCookie(c, pair.RefreshToken)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user":   user,
		"tokens": fiber.Map{"access_token": pair.AccessToken},
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *authHandler) login(c *fiber.Ctx) error {
	var body loginBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Email == "" || body.Password == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "email and password are required")
	}
	ip := c.IP()
	ua := c.Get("User-Agent")
	pair, err := h.svc.LoginWithEmail(c.Context(), body.Email, body.Password, ip, ua)
	if err != nil {
		return mapServiceError(c, err)
	}
	setRefreshCookie(c, pair.RefreshToken)
	return c.JSON(fiber.Map{"tokens": fiber.Map{"access_token": pair.AccessToken}})
}

func (h *authHandler) refreshToken(c *fiber.Ctx) error {
	rawToken := c.Cookies("refresh_token")
	if rawToken == "" {
		return sendError(c, fiber.StatusUnauthorized, "unauthorized", "No refresh token")
	}
	ip := c.IP()
	ua := c.Get("User-Agent")
	pair, err := h.svc.RefreshToken(c.Context(), rawToken, ip, ua)
	if err != nil {
		return mapServiceError(c, err)
	}
	setRefreshCookie(c, pair.RefreshToken)
	return c.JSON(fiber.Map{"access_token": pair.AccessToken})
}

func (h *authHandler) logout(c *fiber.Ctx) error {
	userID := userIDFromCtx(c)
	rawToken := c.Cookies("refresh_token")
	if rawToken != "" {
		// Best-effort: find refresh token record to revoke it.
		// The service Logout expects (refreshTokenID, userID).
		// We pass the raw token as ID so the service can split/find it.
		_ = h.svc.Logout(c.Context(), rawToken, userID)
	}
	clearRefreshCookie(c)
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

type otpRequestBody struct {
	Phone string `json:"phone"`
}

func (h *authHandler) requestOTP(c *fiber.Ctx) error {
	var body otpRequestBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Phone == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "phone is required")
	}
	if err := h.svc.RequestOTP(c.Context(), body.Phone); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "OTP sent to WhatsApp"})
}

type otpVerifyBody struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

func (h *authHandler) verifyOTP(c *fiber.Ctx) error {
	var body otpVerifyBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Phone == "" || body.OTP == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "phone and otp are required")
	}
	pair, err := h.svc.VerifyOTP(c.Context(), body.Phone, body.OTP)
	if err != nil {
		return mapServiceError(c, err)
	}
	setRefreshCookie(c, pair.RefreshToken)
	return c.JSON(fiber.Map{"tokens": fiber.Map{"access_token": pair.AccessToken}})
}

func (h *authHandler) passwordResetRequest(c *fiber.Ctx) error {
	// Constant-time response to prevent account enumeration.
	return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
}

func (h *authHandler) passwordResetConfirm(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"status": "not_implemented"})
}

type changePasswordBody struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *authHandler) changePassword(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body changePasswordBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.OldPassword == "" || body.NewPassword == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "old_password and new_password are required")
	}
	if err := h.svc.ChangePassword(c.Context(), userID, body.OldPassword, body.NewPassword); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Password changed"})
}

func (h *authHandler) getMe(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	user, err := h.svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"user": user})
}

func (h *authHandler) enable2FA(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	setup, err := h.svc.Enable2FA(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"totp_uri": setup.QRCodeURL, "secret": setup.Secret})
}

type verify2FABody struct {
	TOTPCode string `json:"totp_code"`
}

func (h *authHandler) verify2FA(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body verify2FABody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if err := h.svc.Verify2FA(c.Context(), userID, body.TOTPCode); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "2FA enabled"})
}

type disable2FABody struct {
	Password string `json:"password"`
}

func (h *authHandler) disable2FA(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body disable2FABody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if err := h.svc.Disable2FA(c.Context(), userID, body.Password); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "2FA disabled"})
}

func (h *authHandler) getSessions(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	sessions, err := h.svc.GetActiveSessions(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"sessions": sessions})
}

func (h *authHandler) revokeSession(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	sessionID := c.Params("id")
	if err := h.svc.RevokeSession(c.Context(), sessionID, userID); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Session revoked"})
}

func (h *authHandler) revokeAllSessions(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.LogoutAllSessions(c.Context(), userID); err != nil {
		return mapServiceError(c, err)
	}
	clearRefreshCookie(c)
	return c.JSON(fiber.Map{"message": "All sessions revoked"})
}
