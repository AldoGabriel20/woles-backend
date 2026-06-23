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
	ID            string     `json:"id"`
	OwnerUserID   string     `json:"owner_user_id"`
	Name          string     `json:"name"`
	Role          MemberRole `json:"role"`
	RelationLabel *string    `json:"relation_label,omitempty"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
