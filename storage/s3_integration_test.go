//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"arcana-db-backup/storage"
)

// s3ConfigFromEnv reads MinIO/S3 connection settings from env vars.
// Required: S3_ENDPOINT, S3_BUCKET, S3_REGION, S3_ACCESS_KEY, S3_SECRET_KEY.
func s3ConfigFromEnv(t *testing.T) storage.S3Config {
	t.Helper()
	cfg := storage.S3Config{
		Endpoint:  os.Getenv("S3_ENDPOINT"),
		Bucket:    os.Getenv("S3_BUCKET"),
		Region:    os.Getenv("S3_REGION"),
		AccessKey: os.Getenv("S3_ACCESS_KEY"),
		SecretKey: os.Getenv("S3_SECRET_KEY"),
	}
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		t.Skip("integration test requires S3_ENDPOINT/S3_BUCKET/S3_ACCESS_KEY/S3_SECRET_KEY env vars")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return cfg
}

// ensureBucket creates the bucket if it does not yet exist (idempotent).
func ensureBucket(t *testing.T, cfg storage.S3Config) {
	t.Helper()
	ctx := context.Background()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})
	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(cfg.Bucket)})
	if err == nil {
		return
	}
	// Tolerate only "already owned by you".
	var owned *s3types.BucketAlreadyOwnedByYou
	if errors.As(err, &owned) {
		return
	}
	t.Fatalf("CreateBucket: %v", err)
}

// TestUploadDownloadRoundtrip exercises Step 2: push an encrypted-looking
// payload to MinIO via storage.Upload, pull it back with storage.Download,
// and verify the bytes match exactly.
func TestUploadDownloadRoundtrip(t *testing.T) {
	cfg := s3ConfigFromEnv(t)
	ensureBucket(t, cfg)

	dir := t.TempDir()
	src := filepath.Join(dir, "payload.bin")

	payload := make([]byte, 256*1024) // 256 KiB of random bytes
	if _, err := rand.Read(payload); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}
	if err := os.WriteFile(src, payload, 0600); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	if err := storage.Upload(cfg, src); err != nil {
		t.Fatalf("storage.Upload: %v", err)
	}
	// The object key is the local filename (matches main.go behavior).
	key := src

	// Give the backend a brief moment in case of eventual list-after-write.
	time.Sleep(100 * time.Millisecond)

	dst := filepath.Join(dir, "payload.downloaded.bin")
	if err := storage.Download(cfg, key, dst); err != nil {
		t.Fatalf("storage.Download: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read downloaded: %v", err)
	}
	if !bytes.Equal(payload, got) {
		t.Fatalf("downloaded bytes do not match uploaded (orig=%d, got=%d)",
			len(payload), len(got))
	}
}
