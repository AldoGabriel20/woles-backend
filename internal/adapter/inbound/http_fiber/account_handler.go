package http_fiber

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	appidentity "github.com/woles/woles-backend/internal/application/identity"
	"github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/woles/woles-backend/internal/port/outbound/storage"
)

type accountHandler struct {
	svc   *appidentity.Service
	users database.UserRepository
	store storage.FileStore
}

// RegisterAccountRoutes mounts all /api/v1/account routes.
func RegisterAccountRoutes(router fiber.Router, svc *Services) {
	h := &accountHandler{svc: svc.Identity, users: svc.Users, store: svc.FileStore}
	a := router.Group("/account")
	a.Get("/profile", h.getProfile)
	a.Patch("/profile", h.updateProfile)
	a.Post("/avatar", h.uploadAvatar)
	a.Delete("/", h.deleteAccount)
}

func (h *accountHandler) getProfile(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	user, err := h.svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"profile": user})
}

type updateProfileBody struct {
	Name     string `json:"name"`
	Timezone string `json:"timezone"`
}

func (h *accountHandler) updateProfile(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateProfileBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	user, err := h.svc.UpdateProfile(c.Context(), userID, body.Name, body.Timezone)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"profile": user})
}

func (h *accountHandler) uploadAvatar(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	file, err := c.FormFile("avatar")
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "avatar field is required")
	}
	f, err := file.Open()
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, "internal_error", "Failed to open upload")
	}
	defer f.Close()

	key := fmt.Sprintf("avatars/%s/%s", userID, file.Filename)
	objectKey, err := h.store.Upload(c.Context(), key, file.Header.Get("Content-Type"), f)
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, "upload_failed", "Failed to upload avatar")
	}
	if err := h.users.UpdateAvatarURL(c.Context(), userID, objectKey); err != nil {
		return sendError(c, fiber.StatusInternalServerError, "internal_error", "Failed to save avatar URL")
	}
	return c.JSON(fiber.Map{"avatar_url": objectKey})
}

func (h *accountHandler) deleteAccount(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.SoftDeleteUser(c.Context(), userID); err != nil {
		return mapServiceError(c, err)
	}
	clearRefreshCookie(c)
	return c.JSON(fiber.Map{"message": "Account deleted"})
}
