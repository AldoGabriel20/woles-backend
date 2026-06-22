package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/document"
	"github.com/woles/woles-backend/internal/port/outbound/database"
)

// DocumentRepo implements database.DocumentRepository.
type DocumentRepo struct {
	pool *pgxpool.Pool
}

// NewDocumentRepo creates a new DocumentRepo.
func NewDocumentRepo(pool *pgxpool.Pool) *DocumentRepo {
	return &DocumentRepo{pool: pool}
}

var documentSortAllowlist = map[string]struct{}{
	"id": {}, "created_at": {}, "updated_at": {}, "expiry_date": {}, "title": {},
}

const documentSelectCols = `id, user_id, family_member_id, document_type, vault_category,
       title, expiry_date, reminder_offsets, notes, storage_type,
       file_url, file_size_bytes, file_mime_type, status, created_at, updated_at`

func scanDocument(row interface {
	Scan(dest ...any) error
}) (*document.Document, error) {
	d := &document.Document{}
	var docType, vaultCat, storageType, status string
	var offsets []int32
	err := row.Scan(
		&d.ID, &d.UserID, &d.FamilyMemberID,
		&docType, &vaultCat,
		&d.Title, &d.ExpiryDate, &offsets, &d.Notes,
		&storageType,
		&d.FileURL, &d.FileSizeBytes, &d.FileMIMEType,
		&status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.DocumentType = document.DocumentType(docType)
	d.VaultCategory = document.VaultCategory(vaultCat)
	d.StorageType = document.StorageType(storageType)
	d.Status = document.DocumentStatus(status)
	d.ReminderOffsets = make([]int, len(offsets))
	for i, v := range offsets {
		d.ReminderOffsets[i] = int(v)
	}
	return d, nil
}

// Create inserts a new document.
func (r *DocumentRepo) Create(ctx context.Context, d *document.Document) error {
	offsets := make([]int32, len(d.ReminderOffsets))
	for i, v := range d.ReminderOffsets {
		offsets[i] = int32(v)
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO documents
		  (id, user_id, family_member_id, document_type, vault_category,
		   title, expiry_date, reminder_offsets, notes, storage_type,
		   file_url, file_size_bytes, file_mime_type, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		d.ID, d.UserID, d.FamilyMemberID,
		string(d.DocumentType), string(d.VaultCategory),
		d.Title, d.ExpiryDate, offsets, d.Notes,
		string(d.StorageType),
		d.FileURL, d.FileSizeBytes, d.FileMIMEType,
		string(d.Status), d.CreatedAt, d.UpdatedAt,
	)
	return err
}

// FindByID returns the document with the given ID, or ErrNotFound.
func (r *DocumentRepo) FindByID(ctx context.Context, id string) (*document.Document, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM documents WHERE id = $1 AND status != 'archived'`, documentSelectCols),
		id,
	)
	d, err := scanDocument(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return d, err
}

// FindAllByUser returns a paginated list of documents for the user.
func (r *DocumentRepo) FindAllByUser(
	ctx context.Context, userID string,
	filter database.DocumentFilter, p database.PaginationParams,
) (*database.PaginatedResult[*document.Document], error) {
	order := safeOrderBy(p.Sort, p.Order, documentSortAllowlist, "created_at")
	args := []any{userID}
	where := "user_id = $1 AND status != 'archived'"
	idx := 2

	if filter.VaultCategory != nil {
		where += fmt.Sprintf(" AND vault_category = $%d", idx)
		args = append(args, string(*filter.VaultCategory))
		idx++
	}
	if filter.ExpiryWithinDays != nil {
		where += fmt.Sprintf(" AND expiry_date <= NOW() + INTERVAL '%d days'", *filter.ExpiryWithinDays)
	}
	if filter.Search != nil && *filter.Search != "" {
		where += fmt.Sprintf(" AND title ILIKE $%d", idx)
		args = append(args, "%"+*filter.Search+"%")
		idx++
	}

	var total int
	if err := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT COUNT(*) FROM documents WHERE %s`, where), args...,
	).Scan(&total); err != nil {
		return nil, err
	}

	offset := pageOffset(p.Page, p.PerPage)
	args = append(args, p.PerPage, offset)
	query := fmt.Sprintf(`SELECT %s FROM documents WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`,
		documentSelectCols, where, order, idx, idx+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*document.Document
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &database.PaginatedResult[*document.Document]{
		Items: items, Total: total, Page: p.Page,
		PerPage: p.PerPage, TotalPages: totalPages(total, p.PerPage),
	}, nil
}

// Update overwrites a document row.
func (r *DocumentRepo) Update(ctx context.Context, d *document.Document) error {
	offsets := make([]int32, len(d.ReminderOffsets))
	for i, v := range d.ReminderOffsets {
		offsets[i] = int32(v)
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE documents SET
		  document_type = $1, vault_category = $2, title = $3, expiry_date = $4,
		  reminder_offsets = $5, notes = $6, storage_type = $7,
		  file_url = $8, file_size_bytes = $9, file_mime_type = $10,
		  status = $11, updated_at = NOW()
		WHERE id = $12`,
		string(d.DocumentType), string(d.VaultCategory), d.Title, d.ExpiryDate,
		offsets, d.Notes, string(d.StorageType),
		d.FileURL, d.FileSizeBytes, d.FileMIMEType,
		string(d.Status), d.ID,
	)
	return err
}

// SoftDelete archives a document by setting status to 'archived'.
func (r *DocumentRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE documents SET status = 'archived', updated_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// FindExpiringSoon returns active documents expiring within withinDays days.
func (r *DocumentRepo) FindExpiringSoon(ctx context.Context, userID string, withinDays int) ([]*document.Document, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM documents
		WHERE user_id = $1 AND status = 'active'
		  AND expiry_date IS NOT NULL
		  AND expiry_date <= NOW() + INTERVAL '%d days'
		ORDER BY expiry_date ASC`, documentSelectCols, withinDays),
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*document.Document
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

// GetStorageUsage returns total file storage used by the user.
func (r *DocumentRepo) GetStorageUsage(ctx context.Context, userID string) (*database.StorageUsage, error) {
	su := &database.StorageUsage{}
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(file_size_bytes), 0), COUNT(*)
		FROM documents
		WHERE user_id = $1 AND file_url IS NOT NULL`,
		userID,
	).Scan(&su.UsedBytes, &su.FileCount)
	if err != nil {
		return nil, err
	}
	su.UsedMB = float64(su.UsedBytes) / (1024 * 1024)
	su.UsedGB = float64(su.UsedBytes) / (1024 * 1024 * 1024)
	return su, nil
}

// GetVaultHealth returns document counts per vault category and a completeness score.
// Completeness is 100% when the user has at least one doc in vehicles, identity, and insurance.
func (r *DocumentRepo) GetVaultHealth(ctx context.Context, userID string) (*database.VaultHealth, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT vault_category, COUNT(*) AS cnt
		FROM documents
		WHERE user_id = $1 AND status = 'active'
		GROUP BY vault_category`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vh := &database.VaultHealth{}
	catCounts := map[document.VaultCategory]int{}
	for rows.Next() {
		var cat string
		var cnt int
		if err := rows.Scan(&cat, &cnt); err != nil {
			return nil, err
		}
		vc := document.VaultCategory(cat)
		catCounts[vc] = cnt
		vh.Categories = append(vh.Categories, database.VaultCategoryCount{Category: vc, Count: cnt})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Score: vehicles + identity + insurance each worth 33 points; round up to 100 if all present.
	coreCategories := []document.VaultCategory{
		document.VaultVehicles, document.VaultIdentity, document.VaultInsurance,
	}
	met := 0
	for _, c := range coreCategories {
		if catCounts[c] > 0 {
			met++
		}
	}
	switch met {
	case 3:
		vh.CompletenessScore = 100
	case 2:
		vh.CompletenessScore = 66
	case 1:
		vh.CompletenessScore = 33
	default:
		vh.CompletenessScore = 0
	}
	return vh, nil
}
