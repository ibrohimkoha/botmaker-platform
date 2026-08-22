package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrB2NotConfigured is returned when an upload is attempted without a
// valid Backblaze B2 setup.
var ErrB2NotConfigured = errors.New("storage: b2 uploader is not configured")

// B2Uploader uploads files to a Backblaze B2 bucket through its
// S3-compatible API. The client is created lazily; no network I/O
// happens until the first upload.
type B2Uploader struct {
	client *minio.Client
	bucket string
}

// NewB2Uploader builds an uploader for a Backblaze B2 S3 endpoint.
func NewB2Uploader(endpoint, bucket, keyID, appKey, region string) (*B2Uploader, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(keyID, appKey, ""),
		Secure: true,
		Region: region,
	})
	if err != nil {
		return nil, fmt.Errorf("new b2 client: %w", err)
	}
	return &B2Uploader{client: client, bucket: bucket}, nil
}

// Upload streams content under objectName and returns its public URL.
// The object becomes publicly readable at
// https://<endpoint-host>/<bucket>/<objectName>.
func (u *B2Uploader) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (string, error) {
	if u == nil || u.client == nil {
		return "", ErrB2NotConfigured
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := u.client.PutObject(ctx, u.bucket, objectName, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("b2 upload: %w", err)
	}
	return fmt.Sprintf("https://%s/%s/%s", u.client.EndpointURL().Host, u.bucket, objectName), nil
}

// UploadMultipartFile uploads an opened multipart file part and returns
// its public URL.
func (u *B2Uploader) UploadMultipartFile(ctx context.Context, objectName string, f multipart.File, h *multipart.FileHeader) (string, error) {
	return u.Upload(ctx, objectName, f, h.Size, h.Header.Get("Content-Type"))
}
