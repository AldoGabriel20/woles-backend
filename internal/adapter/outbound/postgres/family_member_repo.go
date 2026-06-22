package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/woles/woles-backend/internal/domain/family"
)

// FamilyMemberRepo implements database.FamilyMemberRepository.
type FamilyMemberRepo struct {
	pool *pgxpool.Pool
}

// NewFamilyMemberRepo creates a new FamilyMemberRepo.
func NewFamilyMemberRepo(pool *pgxpool.Pool) *FamilyMemberRepo {
	return &FamilyMemberRepo{pool: pool}
}

const familyMemberSelectCols = `id, owner_user_id, name, role, relation_label, avatar_url, created_at, updated_at`

func scanFamilyMember(row interface {
	Scan(dest ...any) error
}) (*family.FamilyMember, error) {
	m := &family.FamilyMember{}
	var role string
	err := row.Scan(
		&m.ID, &m.OwnerUserID, &m.Name, &role,
		&m.RelationLabel, &m.AvatarURL,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.Role = family.MemberRole(role)
	return m, nil
}

// Create inserts a new family member.
func (r *FamilyMemberRepo) Create(ctx context.Context, m *family.FamilyMember) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO family_members
		  (id, owner_user_id, name, role, relation_label, avatar_url, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.OwnerUserID, m.Name, string(m.Role),
		m.RelationLabel, m.AvatarURL, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

// FindByID returns the family member with the given ID, or ErrNotFound.
func (r *FamilyMemberRepo) FindByID(ctx context.Context, id string) (*family.FamilyMember, error) {
	row := r.pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT %s FROM family_members WHERE id = $1`, familyMemberSelectCols),
		id,
	)
	m, err := scanFamilyMember(row)
	if isNotFound(err) {
		return nil, ErrNotFound
	}
	return m, err
}

// FindAllByOwner returns all family members for the given owner.
func (r *FamilyMemberRepo) FindAllByOwner(ctx context.Context, ownerUserID string) ([]*family.FamilyMember, error) {
	rows, err := r.pool.Query(ctx,
		fmt.Sprintf(`SELECT %s FROM family_members WHERE owner_user_id = $1 ORDER BY created_at ASC`, familyMemberSelectCols),
		ownerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*family.FamilyMember
	for rows.Next() {
		m, err := scanFamilyMember(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// Update overwrites a family member row.
func (r *FamilyMemberRepo) Update(ctx context.Context, m *family.FamilyMember) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE family_members SET
		  name = $1, role = $2, relation_label = $3, avatar_url = $4, updated_at = NOW()
		WHERE id = $5`,
		m.Name, string(m.Role), m.RelationLabel, m.AvatarURL, m.ID,
	)
	return err
}

// Delete hard-deletes a family member row.
func (r *FamilyMemberRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM family_members WHERE id = $1`, id)
	return err
}
