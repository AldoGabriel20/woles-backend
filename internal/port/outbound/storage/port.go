// Package storage defines outbound port interfaces for object storage providers.
package storage

import (
	"context"
	"io"
	"time"
)

// FileStore provides object storage operations backed by MinIO or any S3-compatible provider.
type FileStore interface {
	// Upload streams data to the given key with the specified MIME type.
	// Returns the object key (not a public URL) on success.
	Upload(ctx context.Context, key, mimeType string, data io.Reader) (objectKey string, err error)

	// Delete permanently removes the object at the given key.
	Delete(ctx context.Context, key string) error

	// SignedURL generates a pre-signed GET URL valid for the given duration.
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
