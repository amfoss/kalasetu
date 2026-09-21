// Package storage provides a generic object storage abstraction that backend
// services use to persist binary media (images, videos, ...). The concrete S3
// implementation keeps AWS-specific details out of the service layer so other
// KalaSetu features can reuse the same interface.
package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"kalasetu/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ObjectStorage is the generic contract used by services to upload, delete and
// resolve public URLs of objects. Implementations must be safe for concurrent use.
type ObjectStorage interface {
	// Upload stores the reader's contents under key, tagged with contentType.
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	// Delete removes the object stored under key. Missing objects are not an error.
	Delete(ctx context.Context, key string) error
	// GetURL returns a publicly accessible URL for the object stored under key.
	GetURL(ctx context.Context, key string) (string, error)
}

// S3 is the AWS S3 backed implementation of ObjectStorage.
type S3 struct {
	client          *s3.Client
	bucket          string
	region          string
	publicEndpoint  string
}

// NewS3 builds an S3 client from the given storage configuration. When no
// static credentials are provided it falls back to the AWS default credential
// chain (environment variables, shared config, ECS/EC2 IAM roles, ...).
func NewS3(cfg *config.StorageConfig) (*S3, error) {
	if cfg == nil || cfg.Bucket == "" {
		return nil, fmt.Errorf("object storage is not configured: AWS_BUCKET is required")
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	} else if cfg.AccessKeyID != "" || cfg.SecretAccessKey != "" {
		return nil, fmt.Errorf("object storage requires both AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
	}

	return &S3{
		client:         s3.NewFromConfig(awsCfg, s3Options(cfg)),
		bucket:         cfg.Bucket,
		region:         region,
		publicEndpoint: cfg.PublicEndpoint,
	}, nil
}

// s3Options returns client options for the configured endpoint. A custom
// endpoint (AWS_ENDPOINT, e.g. MinIO) forces path-style addressing.
func s3Options(cfg *config.StorageConfig) func(*s3.Options) {
	return func(o *s3.Options) {
		if cfg.Endpoint == "" {
			return
		}
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(cfg.Endpoint)
	}
}

// Upload implements ObjectStorage.
func (s *S3) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	return err
}

// Delete implements ObjectStorage.
func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetURL implements ObjectStorage. When a public endpoint is configured (e.g.
// MinIO) the object URL is constructed path-style from it; otherwise the
// virtual-hosted-style AWS S3 URL is returned for the public-read bucket.
func (s *S3) GetURL(_ context.Context, key string) (string, error) {
	if s.publicEndpoint != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.publicEndpoint, "/"), s.bucket, key), nil
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}
