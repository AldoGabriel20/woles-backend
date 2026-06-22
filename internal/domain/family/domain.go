package family

import "time"

// MemberRole defines the relationship of a family member to the primary account holder.
type MemberRole string

const (
	RolePrimary MemberRole = "primary"
	RoleSpouse  MemberRole = "spouse"
	RoleParent  MemberRole = "parent"
	RoleChild   MemberRole = "child"
	RoleOther   MemberRole = "other"
)

// FamilyMember represents a person managed under a primary user account.
type FamilyMember struct {
	ID            string
	OwnerUserID   string
	Name          string
	Role          MemberRole
	RelationLabel *string
	AvatarURL     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
