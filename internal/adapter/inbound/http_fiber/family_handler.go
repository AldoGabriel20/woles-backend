package http_fiber

import (
	"github.com/gofiber/fiber/v2"
	appfamily "github.com/woles/woles-backend/internal/application/family"
	domainfamily "github.com/woles/woles-backend/internal/domain/family"
)

type familyHandler struct{ svc *appfamily.Service }

// RegisterFamilyRoutes mounts all /api/v1/family routes.
func RegisterFamilyRoutes(router fiber.Router, svc *Services) {
	h := &familyHandler{svc: svc.Family}
	f := router.Group("/family")
	f.Get("/members", h.list)
	f.Post("/members", h.create)
	f.Get("/members/:id", h.get)
	f.Patch("/members/:id", h.update)
	f.Delete("/members/:id", h.delete)
	f.Get("/reminders", h.sharedReminders)
}

type createMemberBody struct {
	Name          string `json:"name"`
	Role          string `json:"role"`
	RelationLabel string `json:"relation_label"`
	AvatarURL     string `json:"avatar_url"`
}

func (h *familyHandler) create(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body createMemberBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Name == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "name is required")
	}
	req := appfamily.CreateMemberRequest{
		Name: body.Name,
		Role: domainfamily.MemberRole(body.Role),
	}
	if body.RelationLabel != "" {
		req.RelationLabel = &body.RelationLabel
	}
	if body.AvatarURL != "" {
		req.AvatarURL = &body.AvatarURL
	}
	member, err := h.svc.CreateMember(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"member": member})
}

func (h *familyHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	members, err := h.svc.GetMembers(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"members": members})
}

func (h *familyHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	member, err := h.svc.GetMemberByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"member": member})
}

type updateMemberBody struct {
	Name          *string `json:"name"`
	Role          *string `json:"role"`
	RelationLabel *string `json:"relation_label"`
	AvatarURL     *string `json:"avatar_url"`
}

func (h *familyHandler) update(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateMemberBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	req := appfamily.UpdateMemberRequest{
		Name:          body.Name,
		RelationLabel: body.RelationLabel,
		AvatarURL:     body.AvatarURL,
	}
	if body.Role != nil {
		role := domainfamily.MemberRole(*body.Role)
		req.Role = &role
	}
	member, err := h.svc.UpdateMember(c.Context(), userID, c.Params("id"), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"member": member})
}

func (h *familyHandler) delete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteMember(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Member deleted"})
}

func (h *familyHandler) sharedReminders(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)
	result, err := h.svc.GetSharedReminders(c.Context(), userID, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"reminders": result.Items,
		"meta":      paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}
