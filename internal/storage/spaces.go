package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SpacesClient wraps the AWS S3 client for DigitalOcean Spaces
type SpacesClient struct {
	client     *s3.Client
	bucketName string
}

// NewSpacesClient creates a new DigitalOcean Spaces client (S3-compatible)
func NewSpacesClient(ctx context.Context) (*SpacesClient, error) {
	endpoint := os.Getenv("SPACES_ENDPOINT")
	region := os.Getenv("SPACES_REGION")
	accessKey := os.Getenv("SPACES_KEY")
	secretKey := os.Getenv("SPACES_SECRET")
	bucketName := os.Getenv("SPACES_BUCKET")

	// Validate required environment variables
	if endpoint == "" || region == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required Spaces environment variables (SPACES_ENDPOINT, SPACES_REGION, SPACES_KEY, SPACES_SECRET, SPACES_BUCKET)")
	}

	// Create AWS config with static credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Required for DigitalOcean Spaces
	})

	return &SpacesClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to Spaces
// key format: "audio/<user_id>/<uuid>.mp3" or "artwork/<user_id>/<uuid>.jpg"
func (sc *SpacesClient) UploadFile(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := sc.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(sc.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         "private", // Private by default
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to Spaces: %w", err)
	}
	return nil
}

// UploadStream uploads a file from an io.Reader (for streaming uploads)
func (sc *SpacesClient) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := sc.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(sc.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		ACL:         "private",
	})
	if err != nil {
		return fmt.Errorf("failed to upload stream to Spaces: %w", err)
	}
	return nil
}

// CreateSignedURL generates a pre-signed URL for secure file access
// This mirrors Supabase's createSignedUrl functionality
func (sc *SpacesClient) CreateSignedURL(ctx context.Context, key string, expiresInSeconds int) (string, error) {
	presignClient := s3.NewPresignClient(sc.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(sc.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expiresInSeconds) * time.Second
	})
	if err != nil {
		return "", fmt.Errorf("failed to create signed URL: %w", err)
	}

	return req.URL, nil
}

// DeleteFile deletes a file from Spaces
func (sc *SpacesClient) DeleteFile(ctx context.Context, key string) error {
	_, err := sc.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(sc.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from Spaces: %w", err)
	}
	return nil
}

// FileExists checks if a file exists in Spaces
func (sc *SpacesClient) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := sc.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(sc.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if it's a not found error
		return false, nil
	}
	return true, nil
}
