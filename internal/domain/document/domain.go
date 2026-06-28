package document

import (
	"fmt"
	"time"
)

// VaultCategory groups documents by life admin category.
type VaultCategory string

const (
	VaultVehicles   VaultCategory = "vehicles"
	VaultIdentity   VaultCategory = "identity"
	VaultInsurance  VaultCategory = "insurance"
	VaultFinancials VaultCategory = "financials"
	VaultFinancial  VaultCategory = "financial"
	VaultProperty   VaultCategory = "property"
	VaultHealth     VaultCategory = "health"
	VaultEducation  VaultCategory = "education"
	VaultLegal      VaultCategory = "legal"
	VaultOther      VaultCategory = "other"
)

// DocumentType enumerates supported document kinds.
type DocumentType string

const (
	DocTypeSTNK             DocumentType = "stnk"
	DocTypeBPKB             DocumentType = "bpkb"
	DocTypeVehicleInsurance DocumentType = "vehicle_insurance"
	DocTypeSIM              DocumentType = "sim"
	DocTypePassport         DocumentType = "passport"
	DocTypeVisa             DocumentType = "visa"
	DocTypeKTP              DocumentType = "ktp"
	DocTypeHealthInsurance  DocumentType = "health_insurance"
	DocTypeLifeInsurance    DocumentType = "life_insurance"
	DocTypeTax              DocumentType = "tax"
	DocTypeInvestment       DocumentType = "investment"
	DocTypeOther            DocumentType = "other"
)

// StorageType indicates how the document is stored.
type StorageType string

const (
	StoragePhysical     StorageType = "physical"
	StorageDigital      StorageType = "digital"
	StorageScanVerified StorageType = "scan_verified"
)

// DocumentStatus represents the lifecycle state of a document.
type DocumentStatus string

const (
	DocumentStatusActive   DocumentStatus = "active"
	DocumentStatusArchived DocumentStatus = "archived"
)

// Document is the core document vault entity.
type Document struct {
	ID              string         `json:"id"`
	UserID          string         `json:"user_id"`
	FamilyMemberID  *string        `json:"family_member_id,omitempty"`
	DocumentType    DocumentType   `json:"document_type"`
	VaultCategory   VaultCategory  `json:"vault_category"`
	Title           string         `json:"title"`
	ExpiryDate      *time.Time     `json:"expiry_date,omitempty"`
	ReminderOffsets []int          `json:"reminder_offsets,omitempty"`
	Notes           *string        `json:"notes,omitempty"`
	StorageType     StorageType    `json:"storage_type"`
	FileURL         *string        `json:"file_url,omitempty"`
	FileSizeBytes   *int64         `json:"file_size_bytes,omitempty"`
	FileMIMEType    *string        `json:"file_mime_type,omitempty"`
	Status          DocumentStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// DaysUntilExpiry returns the number of calendar days between today and
// ExpiryDate. Returns a large positive number (math.MaxInt32) when there is no
// expiry date, and a negative number when the document is already expired.
func (d *Document) DaysUntilExpiry() int {
	if d.ExpiryDate == nil {
		return 1<<31 - 1 // no expiry — treat as very far future
	}
	now := time.Now().UTC().Truncate(24 * time.Hour)
	expiry := d.ExpiryDate.UTC().Truncate(24 * time.Hour)
	return int(expiry.Sub(now).Hours() / 24)
}

// ExpiryRisk categorises how close the document is to expiry.
//
//   - "expired"  — already expired
//   - "urgent"   — expires within 7 days
//   - "upcoming" — expires within 30 days
//   - "safe"     — more than 30 days remaining (or no expiry date)
func (d *Document) ExpiryRisk() string {
	days := d.DaysUntilExpiry()
	switch {
	case days < 0:
		return "expired"
	case days <= 7:
		return "urgent"
	case days <= 30:
		return "upcoming"
	default:
		return "safe"
	}
}

// VaultCategoryForType derives the vault category from a document type.
func VaultCategoryForType(dt DocumentType) VaultCategory {
	switch dt {
	case DocTypeSTNK, DocTypeBPKB, DocTypeVehicleInsurance:
		return VaultVehicles
	case DocTypeSIM, DocTypePassport, DocTypeVisa, DocTypeKTP:
		return VaultIdentity
	case DocTypeHealthInsurance, DocTypeLifeInsurance:
		return VaultInsurance
	case DocTypeTax, DocTypeInvestment:
		return VaultFinancials
	default:
		return VaultOther
	}
}

// StorageKeyFor generates the object-storage key for a document file.
// ext should include the leading dot, e.g. ".pdf".
func StorageKeyFor(userID, documentID, fileUUID, ext string) string {
	return fmt.Sprintf("documents/%s/%s/%s%s", userID, documentID, fileUUID, ext)
}
