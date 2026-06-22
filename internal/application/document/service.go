// Package document implements the Document application service.
package document

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domaindocument "github.com/woles/woles-backend/internal/domain/document"
	domainnotification "github.com/woles/woles-backend/internal/domain/notification"
	"github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/woles/woles-backend/internal/port/outbound/storage"
)

// ─── Errors ───────────────────────────────────────────────────────────────────

var (
	ErrNotFound        = errors.New("document not found")
	ErrForbidden       = errors.New("forbidden")
	ErrPlanLimit       = errors.New("document limit reached for your plan")
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidMIMEType = errors.New("unsupported file type: only PDF, JPEG and PNG are allowed")
	ErrFileTooLarge    = errors.New("file exceeds the 10 MB limit")
	ErrNoFileAttached  = errors.New("no file attached to this document")
)

// ─── Request / response types ─────────────────────────────────────────────────

// CreateDocumentRequest holds the input for creating a new document.
type CreateDocumentRequest struct {
	DocumentType    domaindocument.DocumentType
	VaultCategory   *domaindocument.VaultCategory // optional; derived from DocumentType when absent
	Title           string
	ExpiryDate      *time.Time
	ReminderOffsets []int // default [30, 7, 1] when nil
	Notes           *string
	FamilyMemberID  *string
}

// UpdateDocumentRequest holds the fields that may be updated.
type UpdateDocumentRequest struct {
	Title           *string
	ExpiryDate      *time.Time
	ReminderOffsets []int
	Notes           *string
	FamilyMemberID  *string
}

// StorageUsage contains aggregated file storage metrics for a user.
type StorageUsage = database.StorageUsage

// VaultHealth contains per-category counts and an overall completeness score.
type VaultHealth = database.VaultHealth

// ─── Service ──────────────────────────────────────────────────────────────────

// Service implements the document application service.
type Service struct {
	documents     database.DocumentRepository
	notifications database.NotificationRepository
	usageLimits   database.UsageLimitRepository
	auditLogs     database.AuditLogRepository
	fileStore     storage.FileStore
}

// NewService constructs the document service.
func NewService(
	documents database.DocumentRepository,
	notifications database.NotificationRepository,
	usageLimits database.UsageLimitRepository,
	auditLogs database.AuditLogRepository,
	fileStore storage.FileStore,
) *Service {
	return &Service{
		documents:     documents,
		notifications: notifications,
		usageLimits:   usageLimits,
		auditLogs:     auditLogs,
		fileStore:     fileStore,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

// CreateDocument validates the request, enforces plan limits, and inserts a new
// document together with its expiry notifications.
func (s *Service) CreateDocument(ctx context.Context, userID string, req CreateDocumentRequest) (*domaindocument.Document, error) {
	// Plan limit check.
	within, err := s.usageLimits.IsWithinLimit(ctx, userID, "documents")
	if err != nil {
		return nil, fmt.Errorf("check usage limit: %w", err)
	}
	if !within {
		return nil, ErrPlanLimit
	}

	// Validate document type.
	if !validDocumentType(req.DocumentType) {
		return nil, fmt.Errorf("%w: unknown document_type %q", ErrInvalidInput, req.DocumentType)
	}

	// Validate expiry_date when provided.
	if req.ExpiryDate != nil && req.ExpiryDate.IsZero() {
		return nil, fmt.Errorf("%w: expiry_date is not a valid date", ErrInvalidInput)
	}

	// Sanitize strings.
	title := sanitize(req.Title, 200)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	var notes *string
	if req.Notes != nil {
		n := sanitize(*req.Notes, 2000)
		notes = &n
	}

	// Derive vault_category from document_type when not explicitly provided.
	vaultCategory := domaindocument.VaultCategoryForType(req.DocumentType)
	if req.VaultCategory != nil {
		vaultCategory = *req.VaultCategory
	}

	// Default reminder offsets.
	offsets := req.ReminderOffsets
	if len(offsets) == 0 {
		offsets = []int{30, 7, 1}
	}

	now := time.Now().UTC()
	doc := &domaindocument.Document{
		ID:              uuid.NewString(),
		UserID:          userID,
		FamilyMemberID:  req.FamilyMemberID,
		DocumentType:    req.DocumentType,
		VaultCategory:   vaultCategory,
		Title:           title,
		ExpiryDate:      req.ExpiryDate,
		ReminderOffsets: offsets,
		Notes:           notes,
		StorageType:     domaindocument.StoragePhysical, // default; updated on file upload
		Status:          domaindocument.DocumentStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.documents.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	_ = s.usageLimits.Increment(ctx, userID, "documents")

	// Schedule expiry notifications for each offset (only if expiry_date is set
	// and the target date is still in the future).
	if doc.ExpiryDate != nil {
		s.createExpiryNotifications(ctx, userID, doc)
	}

	s.writeAudit(ctx, userID, "document", "document_created", doc.ID)

	return doc, nil
}

// ─── File upload ──────────────────────────────────────────────────────────────

// allowedMIMETypes contains the set of MIME types accepted for document files.
var allowedMIMETypes = map[string]string{
	"application/pdf": "pdf",
	"image/jpeg":      "jpg",
	"image/png":       "png",
}

const maxFileSizeBytes = 10 * 1024 * 1024 // 10 MB

// UploadDocumentFile validates, stores, and links a file to an existing document.
func (s *Service) UploadDocumentFile(ctx context.Context, userID, documentID string, file io.Reader, mimeType string, sizeBytes int) (*domaindocument.Document, error) {
	doc, err := s.ownershipCheck(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	// Validate MIME type.
	ext, ok := allowedMIMETypes[mimeType]
	if !ok {
		return nil, ErrInvalidMIMEType
	}

	// Validate size.
	if sizeBytes > maxFileSizeBytes {
		return nil, ErrFileTooLarge
	}

	// Generate a deterministic, collision-free storage key.
	fileKey := fmt.Sprintf("documents/%s/%s/%s.%s", userID, documentID, uuid.NewString(), ext)

	objectKey, err := s.fileStore.Upload(ctx, fileKey, mimeType, file)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	now := time.Now().UTC()
	sz := int64(sizeBytes)
	doc.FileURL = &objectKey
	doc.FileSizeBytes = &sz
	doc.FileMIMEType = &mimeType
	doc.StorageType = domaindocument.StorageDigital
	doc.UpdatedAt = now

	if err := s.documents.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("update document after upload: %w", err)
	}

	s.writeAudit(ctx, userID, "document", "document_file_uploaded", doc.ID)

	return doc, nil
}

// ─── Signed URL ───────────────────────────────────────────────────────────────

// GetDocumentFileURL returns a 15-minute pre-signed download URL for the file
// attached to a document. Returns ErrNoFileAttached when no file is stored.
func (s *Service) GetDocumentFileURL(ctx context.Context, userID, documentID string) (string, error) {
	doc, err := s.ownershipCheck(ctx, userID, documentID)
	if err != nil {
		return "", err
	}

	if doc.FileURL == nil || *doc.FileURL == "" {
		return "", ErrNoFileAttached
	}

	url, err := s.fileStore.SignedURL(ctx, *doc.FileURL, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("generate signed url: %w", err)
	}

	return url, nil
}

// ─── Delete file ──────────────────────────────────────────────────────────────

// DeleteDocumentFile removes the stored file, clears file-related fields on the
// document, and writes an audit log.
func (s *Service) DeleteDocumentFile(ctx context.Context, userID, documentID string) error {
	doc, err := s.ownershipCheck(ctx, userID, documentID)
	if err != nil {
		return err
	}

	if doc.FileURL == nil || *doc.FileURL == "" {
		return ErrNoFileAttached
	}

	if err := s.fileStore.Delete(ctx, *doc.FileURL); err != nil {
		return fmt.Errorf("delete file from storage: %w", err)
	}

	now := time.Now().UTC()
	doc.FileURL = nil
	doc.FileSizeBytes = nil
	doc.FileMIMEType = nil
	doc.StorageType = domaindocument.StoragePhysical
	doc.UpdatedAt = now

	if err := s.documents.Update(ctx, doc); err != nil {
		return fmt.Errorf("update document after file deletion: %w", err)
	}

	s.writeAudit(ctx, userID, "document", "document_file_deleted", doc.ID)

	return nil
}

// ─── Read ─────────────────────────────────────────────────────────────────────

// GetDocuments returns a paginated list of documents for a user.
func (s *Service) GetDocuments(ctx context.Context, userID string, filter database.DocumentFilter, page, perPage int) (*database.PaginatedResult[*domaindocument.Document], error) {
	p := database.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "created_at",
		Order:   "desc",
	}
	return s.documents.FindAllByUser(ctx, userID, filter, p)
}

// GetDocumentByID returns a single document, enforcing ownership.
func (s *Service) GetDocumentByID(ctx context.Context, userID, documentID string) (*domaindocument.Document, error) {
	return s.ownershipCheck(ctx, userID, documentID)
}

// ─── Update ───────────────────────────────────────────────────────────────────

// UpdateDocument applies partial updates to a document. When expiry_date
// changes, all previously-scheduled expiry notifications are canceled and new
// ones are created for the updated date.
func (s *Service) UpdateDocument(ctx context.Context, userID, documentID string, req UpdateDocumentRequest) (*domaindocument.Document, error) {
	doc, err := s.ownershipCheck(ctx, userID, documentID)
	if err != nil {
		return nil, err
	}

	expiryChanged := false

	if req.Title != nil {
		doc.Title = sanitize(*req.Title, 200)
	}
	if req.ExpiryDate != nil {
		expiryChanged = true
		doc.ExpiryDate = req.ExpiryDate
	}
	if req.ReminderOffsets != nil {
		doc.ReminderOffsets = req.ReminderOffsets
		if doc.ExpiryDate != nil {
			expiryChanged = true // recalculate notifications even if date itself didn't change
		}
	}
	if req.Notes != nil {
		n := sanitize(*req.Notes, 2000)
		doc.Notes = &n
	}
	if req.FamilyMemberID != nil {
		doc.FamilyMemberID = req.FamilyMemberID
	}

	doc.UpdatedAt = time.Now().UTC()

	if err := s.documents.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("update document: %w", err)
	}

	if expiryChanged {
		s.cancelExpiryNotifications(ctx, doc.ID)
		if doc.ExpiryDate != nil {
			s.createExpiryNotifications(ctx, userID, doc)
		}
	}

	return doc, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

// DeleteDocument soft-deletes a document, cancels pending notifications, removes
// any stored file, and decrements the usage counter.
func (s *Service) DeleteDocument(ctx context.Context, userID, documentID string) error {
	doc, err := s.ownershipCheck(ctx, userID, documentID)
	if err != nil {
		return err
	}

	if err := s.documents.SoftDelete(ctx, doc.ID); err != nil {
		return fmt.Errorf("soft delete document: %w", err)
	}

	s.cancelExpiryNotifications(ctx, doc.ID)

	// Remove stored file if one is attached.
	if doc.FileURL != nil && *doc.FileURL != "" {
		_ = s.fileStore.Delete(ctx, *doc.FileURL)
	}

	_ = s.usageLimits.Decrement(ctx, userID, "documents")
	s.writeAudit(ctx, userID, "document", "document_deleted", doc.ID)

	return nil
}

// ─── Aggregates ───────────────────────────────────────────────────────────────

// GetStorageUsage returns the total file storage consumed by the user's
// documents (files only; documents with no attached file are excluded).
func (s *Service) GetStorageUsage(ctx context.Context, userID string) (*StorageUsage, error) {
	usage, err := s.documents.GetStorageUsage(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get storage usage: %w", err)
	}
	return usage, nil
}

// GetVaultHealth returns per-category document counts and an overall
// completeness score (0–100).
func (s *Service) GetVaultHealth(ctx context.Context, userID string) (*VaultHealth, error) {
	health, err := s.documents.GetVaultHealth(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get vault health: %w", err)
	}
	return health, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// ownershipCheck fetches a document by ID and verifies it belongs to userID.
// Returns ErrNotFound (opaque) on missing or unowned documents.
func (s *Service) ownershipCheck(ctx context.Context, userID, documentID string) (*domaindocument.Document, error) {
	doc, err := s.documents.FindByID(ctx, documentID)
	if err != nil || doc == nil {
		return nil, ErrNotFound
	}
	if doc.UserID != userID {
		return nil, ErrNotFound // intentionally opaque to prevent enumeration
	}
	return doc, nil
}

// createExpiryNotifications schedules one notification per offset day before
// the document's expiry date. Offsets in the past are silently skipped.
func (s *Service) createExpiryNotifications(ctx context.Context, userID string, doc *domaindocument.Document) {
	if doc.ExpiryDate == nil {
		return
	}
	now := time.Now().UTC()
	for _, offset := range doc.ReminderOffsets {
		scheduledAt := doc.ExpiryDate.UTC().AddDate(0, 0, -offset)
		if !scheduledAt.After(now) {
			continue // skip dates already in the past
		}
		notifID := uuid.NewString()
		n := &domainnotification.Notification{
			ID:             notifID,
			UserID:         userID,
			EntityType:     domainnotification.EntityDocument,
			EntityID:       doc.ID,
			Channel:        domainnotification.ChannelWhatsApp,
			ScheduledAt:    scheduledAt,
			Status:         domainnotification.StatusScheduled,
			IdempotencyKey: fmt.Sprintf("document:%s:offset:%d:channel:whatsapp", doc.ID, offset),
			RetryCount:     0,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_ = s.notifications.Create(ctx, n)
	}
}

// cancelExpiryNotifications sets all scheduled notifications for a document to
// "canceled". Errors are best-effort.
func (s *Service) cancelExpiryNotifications(ctx context.Context, documentID string) {
	entityType := domainnotification.EntityDocument
	status := domainnotification.StatusScheduled
	filter := database.NotificationFilter{
		EntityType: &entityType,
		Status:     &status,
	}
	p := database.PaginationParams{Page: 1, PerPage: 100, Sort: "scheduled_at", Order: "asc"}
	result, err := s.notifications.FindAllByUser(ctx, "", filter, p)
	if err != nil {
		return
	}
	for _, n := range result.Items {
		if n.EntityID == documentID {
			_ = s.notifications.UpdateStatus(ctx, n.ID, domainnotification.StatusCanceled)
		}
	}
}

// writeAudit creates an audit log entry; errors are silently discarded.
func (s *Service) writeAudit(ctx context.Context, userID, entityType, action, entityID string) {
	log := &database.AuditLog{
		ID:         uuid.NewString(),
		UserID:     strPtr(userID),
		ActorType:  "user",
		Action:     action,
		EntityType: strPtr(entityType),
		EntityID:   strPtr(entityID),
		CreatedAt:  time.Now().UTC(),
	}
	_ = s.auditLogs.Create(ctx, log)
}

// validDocumentType returns true when dt is one of the defined enum values.
func validDocumentType(dt domaindocument.DocumentType) bool {
	switch dt {
	case domaindocument.DocTypeSTNK,
		domaindocument.DocTypeBPKB,
		domaindocument.DocTypeVehicleInsurance,
		domaindocument.DocTypeSIM,
		domaindocument.DocTypePassport,
		domaindocument.DocTypeVisa,
		domaindocument.DocTypeKTP,
		domaindocument.DocTypeHealthInsurance,
		domaindocument.DocTypeLifeInsurance,
		domaindocument.DocTypeTax,
		domaindocument.DocTypeInvestment,
		domaindocument.DocTypeOther:
		return true
	}
	return false
}

// sanitize strips whitespace, removes HTML-like tags, and truncates to maxRunes.
func sanitize(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	s = b.String()
	if utf8.RuneCountInString(s) > maxRunes {
		runes := []rune(s)
		s = string(runes[:maxRunes])
	}
	return s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
