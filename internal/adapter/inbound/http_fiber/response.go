package http_fiber

import (
	"errors"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/woles/woles-backend/internal/adapter/outbound/postgres"
	appfamily "github.com/woles/woles-backend/internal/application/family"
	appgoal "github.com/woles/woles-backend/internal/application/goal"
	appidentity "github.com/woles/woles-backend/internal/application/identity"
)

// sendError writes a unified JSON error response.
func sendError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   code,
		"message": message,
	})
}

// userIDFromCtx reads the userID set by JWTMiddleware. Returns "" on missing.
func userIDFromCtx(c *fiber.Ctx) string {
	id, _ := c.Locals("userID").(string)
	return id
}

// requireUserID reads userID from context and sends 401 if missing.
// Returns ("", true) to signal the handler should return immediately.
func requireUserID(c *fiber.Ctx) (string, bool) {
	id := userIDFromCtx(c)
	if id == "" {
		_ = sendError(c, fiber.StatusUnauthorized, "unauthorized", "Token is expired or invalid")
		return "", true
	}
	return id, false
}

// mapServiceError maps well-known application errors to HTTP status codes.
func mapServiceError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		return sendError(c, fiber.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, appidentity.ErrNotFound):
		return sendError(c, fiber.StatusNotFound, "not_found", "Resource not found")
	case errors.Is(err, appidentity.ErrForbidden):
		return sendError(c, fiber.StatusForbidden, "forbidden", "Access denied")
	case errors.Is(err, appidentity.ErrInvalidCredentials):
		return sendError(c, fiber.StatusBadRequest, "invalid_credentials", err.Error())
	case errors.Is(err, appidentity.ErrAccountLocked):
		return sendError(c, fiber.StatusBadRequest, "account_locked", err.Error())
	case errors.Is(err, appidentity.ErrTokenReused):
		return sendError(c, fiber.StatusUnauthorized, "token_reused", err.Error())
	case errors.Is(err, appidentity.ErrTokenInvalid):
		return sendError(c, fiber.StatusUnauthorized, "token_invalid", err.Error())
	case errors.Is(err, appidentity.Err2FANotEnabled):
		return sendError(c, fiber.StatusBadRequest, "setup_required", err.Error())
	case errors.Is(err, appfamily.ErrPlanGate):
		return sendError(c, fiber.StatusForbidden, "plan_required", err.Error())
	case errors.Is(err, appfamily.ErrMemberLimit):
		return sendError(c, fiber.StatusBadRequest, "member_limit", err.Error())
	case errors.Is(err, appgoal.ErrPlanRequired):
		return sendError(c, fiber.StatusForbidden, "plan_required", err.Error())
	case errors.Is(err, appgoal.ErrNotFound):
		return sendError(c, fiber.StatusNotFound, "not_found", "Resource not found")
	default:
		// Check for plan_limit and quota errors by message
		msg := err.Error()
		if contains(msg, "plan limit") || contains(msg, "quota_exceeded") || contains(msg, "quota exceeded") ||
			contains(msg, "limit reached") || contains(msg, "reached for your plan") ||
			contains(msg, "requires a premium or advanced plan") {
			return sendError(c, fiber.StatusForbidden, "plan_limit", msg)
		}
		if contains(msg, "invalid input") {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", msg)
		}
		if contains(msg, "unsupported file type") {
			return sendError(c, fiber.StatusUnsupportedMediaType, "unsupported_media_type", msg)
		}
		if contains(msg, "file exceeds") {
			return sendError(c, fiber.StatusRequestEntityTooLarge, "file_too_large", msg)
		}
		if contains(msg, "forbidden") || contains(msg, "access denied") {
			return sendError(c, fiber.StatusForbidden, "forbidden", "Access denied")
		}
		if contains(msg, "not found") || contains(msg, "no rows") {
			return sendError(c, fiber.StatusNotFound, "not_found", "Resource not found")
		}
		logError(err)
		return sendError(c, fiber.StatusInternalServerError, "internal_error", "An unexpected error occurred")
	}
}

// isDevelopment returns true when APP_ENV is not "production".
func isDevelopment() bool { return os.Getenv("APP_ENV") != "production" }

// logError logs service errors that map to 500 in development.
func logError(err error) {
	if os.Getenv("APP_ENV") != "production" {
		log.Printf("[ERROR] %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && indexStr(s, substr) >= 0)
}

func indexStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// paginationMeta builds the meta object for paginated responses.
func paginationMeta(page, perPage, total, totalPages int) fiber.Map {
	return fiber.Map{
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": totalPages,
	}
}

// pageParam reads "page" query param, default 1, min 1.
func pageParam(c *fiber.Ctx) int {
	p := c.QueryInt("page", 1)
	if p < 1 {
		p = 1
	}
	return p
}

// perPageParam reads "per_page" query param, default defaultVal, max maxVal.
func perPageParam(c *fiber.Ctx, defaultVal, maxVal int) int {
	pp := c.QueryInt("per_page", defaultVal)
	if pp < 1 {
		pp = defaultVal
	}
	if pp > maxVal {
		pp = maxVal
	}
	return pp
}
