// Package storageprovider implements the storage.FileStore port using MinIO (S3-compatible).
package storageprovider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStore implements port/outbound/storage.FileStore.
// The bucket is private; all access is via pre-signed URLs.
type MinIOStore struct {
	client *minio.Client
	bucket string
}

// New creates a MinIOStore by reading connection settings from environment variables:
//   - MINIO_ENDPOINT  (e.g. "localhost:9000")
//   - MINIO_ACCESS_KEY
//   - MINIO_SECRET_KEY
//   - MINIO_BUCKET    (e.g. "woles-documents")
//   - MINIO_USE_SSL   ("true" / "false", default false)
func New(ctx context.Context) (*MinIOStore, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, errors.New("minio: MINIO_ENDPOINT is not set")
	}
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	if bucket == "" {
		return nil, errors.New("minio: MINIO_BUCKET is not set")
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: new client: %w", err)
	}

	// Ensure bucket exists; create it if not.
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio: check bucket: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio: make bucket %s: %w", bucket, err)
		}
	}

	return &MinIOStore{client: client, bucket: bucket}, nil
}

// Upload streams data to key with the given MIME type.
// Returns the object key (not a URL) so the caller can generate signed URLs later.
func (s *MinIOStore) Upload(ctx context.Context, key, mimeType string, data io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, data, -1, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return "", fmt.Errorf("minio: upload %s: %w", key, err)
	}
	return key, nil
}

// Delete permanently removes the object at key.
func (s *MinIOStore) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio: delete %s: %w", key, err)
	}
	return nil
}

// SignedURL generates a pre-signed GET URL valid for ttl.
// The bucket policy is private, so callers must always use this method for downloads.
func (s *MinIOStore) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	reqParams := url.Values{}
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, reqParams)
	if err != nil {
		return "", fmt.Errorf("minio: signed url %s: %w", key, err)
	}
	return presigned.String(), nil
}
