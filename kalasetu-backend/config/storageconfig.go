package config

import "os"

// StorageConfig holds the settings needed to talk to the object storage
// backend (AWS S3, or any S3-compatible store such as MinIO).
type StorageConfig struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	// Endpoint overrides the S3 endpoint (e.g. a local MinIO server).
	Endpoint string
	// PublicEndpoint is used to build publicly accessible URLs; when empty
	// the virtual-hosted-style AWS URL is used instead.
	PublicEndpoint string
}

// LoadStorageConfig reads object storage configuration from environment
// variables. It follows the same pattern as LoadDBConfig.
func LoadStorageConfig() *StorageConfig {
	return &StorageConfig{
		Region:          os.Getenv("AWS_REGION"),
		Bucket:          os.Getenv("AWS_BUCKET"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("AWS_ENDPOINT"),
		PublicEndpoint:  os.Getenv("AWS_PUBLIC_ENDPOINT"),
	}
}

// IsConfigured reports whether a bucket (the minimum required setting) was
// provided. Credentials may also be picked up from the AWS credential chain,
// e.g. IAM roles in ECS/EC2.
func (c *StorageConfig) IsConfigured() bool {
	return c != nil && c.Bucket != ""
}
