package http_fiber

import (
	"time"

	"github.com/gofiber/fiber/v2"
	appdocument "github.com/woles/woles-backend/internal/application/document"
	domaindocument "github.com/woles/woles-backend/internal/domain/document"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

type documentHandler struct{ svc *appdocument.Service }

// RegisterDocumentRoutes mounts all /api/v1/documents routes.
func RegisterDocumentRoutes(router fiber.Router, svc *Services) {
	h := &documentHandler{svc: svc.Document}
	d := router.Group("/documents")

	d.Post("/", h.create)
	d.Get("/", h.list)

	// Static sub-paths before /:id catch-all.
	d.Get("/storage/usage", h.storageUsage)
	d.Get("/vault/health", h.vaultHealth)

	d.Get("/:id", h.get)
	d.Patch("/:id", h.update)
	d.Delete("/:id", h.delete)
	d.Post("/:id/file", h.uploadFile)
	d.Get("/:id/file/url", h.getFileURL)
	d.Delete("/:id/file", h.deleteFile)
}

type createDocumentBody struct {
	Title           string `json:"title"`
	DocumentType    string `json:"document_type"`
	VaultCategory   string `json:"vault_category"`
	ExpiryDate      string `json:"expiry_date"`
	ReminderOffsets []int  `json:"reminder_offsets"`
	Notes           string `json:"notes"`
	FamilyMemberID  string `json:"family_member_id"`
}

func (h *documentHandler) create(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body createDocumentBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	if body.Title == "" {
		return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "title is required")
	}
	req := appdocument.CreateDocumentRequest{
		Title:           body.Title,
		DocumentType:    domaindocument.DocumentType(body.DocumentType),
		ReminderOffsets: body.ReminderOffsets,
	}
	if req.DocumentType == "" {
		req.DocumentType = domaindocument.DocTypeOther
	}
	if body.VaultCategory != "" {
		vc := domaindocument.VaultCategory(body.VaultCategory)
		req.VaultCategory = &vc
	}
	if body.ExpiryDate != "" {
		t, err := time.Parse("2006-01-02", body.ExpiryDate)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "expiry_date must be YYYY-MM-DD")
		}
		req.ExpiryDate = &t
	}
	if body.Notes != "" {
		req.Notes = &body.Notes
	}
	if body.FamilyMemberID != "" {
		req.FamilyMemberID = &body.FamilyMemberID
	}
	doc, err := h.svc.CreateDocument(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"document": doc})
}

func (h *documentHandler) list(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	page := pageParam(c)
	perPage := perPageParam(c, 20, 100)

	filter := database.DocumentFilter{}
	if cat := c.Query("vault_category"); cat != "" {
		vc := domaindocument.VaultCategory(cat)
		filter.VaultCategory = &vc
	}
	if s := c.Query("search"); s != "" {
		filter.Search = &s
	}

	result, err := h.svc.GetDocuments(c.Context(), userID, filter, page, perPage)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{
		"documents": result.Items,
		"meta":      paginationMeta(result.Page, result.PerPage, result.Total, result.TotalPages),
	})
}

func (h *documentHandler) get(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	doc, err := h.svc.GetDocumentByID(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"document": doc})
}

type updateDocumentBody struct {
	Title           *string `json:"title"`
	ExpiryDate      *string `json:"expiry_date"`
	ReminderOffsets []int   `json:"reminder_offsets"`
	Notes           *string `json:"notes"`
	FamilyMemberID  *string `json:"family_member_id"`
}

func (h *documentHandler) update(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	var body updateDocumentBody
	if err := c.BodyParser(&body); err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "Invalid request body")
	}
	req := appdocument.UpdateDocumentRequest{
		Title:           body.Title,
		Notes:           body.Notes,
		ReminderOffsets: body.ReminderOffsets,
		FamilyMemberID:  body.FamilyMemberID,
	}
	if body.ExpiryDate != nil {
		t, err := time.Parse("2006-01-02", *body.ExpiryDate)
		if err != nil {
			return sendError(c, fiber.StatusUnprocessableEntity, "validation_error", "expiry_date must be YYYY-MM-DD")
		}
		req.ExpiryDate = &t
	}
	doc, err := h.svc.UpdateDocument(c.Context(), userID, c.Params("id"), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"document": doc})
}

func (h *documentHandler) delete(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteDocument(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Document deleted"})
}

func (h *documentHandler) uploadFile(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	file, err := c.FormFile("file")
	if err != nil {
		return sendError(c, fiber.StatusBadRequest, "bad_request", "file field is required")
	}
	f, err := file.Open()
	if err != nil {
		return sendError(c, fiber.StatusInternalServerError, "internal_error", "Failed to open upload")
	}
	defer f.Close()
	doc, err := h.svc.UploadDocumentFile(c.Context(), userID, c.Params("id"), f, file.Header.Get("Content-Type"), int(file.Size))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"document": doc})
}

func (h *documentHandler) deleteFile(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	if err := h.svc.DeleteDocumentFile(c.Context(), userID, c.Params("id")); err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"message": "File deleted"})
}

func (h *documentHandler) getFileURL(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	url, err := h.svc.GetDocumentFileURL(c.Context(), userID, c.Params("id"))
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"url": url})
}

func (h *documentHandler) storageUsage(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	usage, err := h.svc.GetStorageUsage(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"storage": usage})
}

func (h *documentHandler) vaultHealth(c *fiber.Ctx) error {
	userID, abort := requireUserID(c)
	if abort {
		return nil
	}
	health, err := h.svc.GetVaultHealth(c.Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(fiber.Map{"health": health})
}
