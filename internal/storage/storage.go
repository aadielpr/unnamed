package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Store is the S3-compatible object storage abstraction used by the app.
type Store interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	URL(key string) string
}

// Config holds the settings needed to build an S3 client.
type Config struct {
	Endpoint        string
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	PublicURLBase   string
}

// S3Store implements Store using the AWS SDK for S3-compatible services.
type S3Store struct {
	client *s3.Client
	bucket string
	urlBase string
}

// NewS3Store creates a Store backed by any S3-compatible service.
func NewS3Store(cfg Config) (Store, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithBaseEndpoint(cfg.Endpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	urlBase := cfg.PublicURLBase
	if urlBase == "" {
		urlBase = cfg.Endpoint
	}

	return &S3Store{
		client:  client,
		bucket:  cfg.Bucket,
		urlBase: urlBase,
	}, nil
}

// Put uploads an object to the configured bucket.
func (s *S3Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Get downloads an object from the configured bucket.
func (s *S3Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return out.Body, nil
}

// Delete removes an object from the configured bucket.
func (s *S3Store) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

// Exists checks whether an object is present in the configured bucket.
func (s *S3Store) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("head object %q: %w", key, err)
	}
	return true, nil
}

// URL returns a public URL for the given object key.
func (s *S3Store) URL(key string) string {
	u, err := url.Parse(s.urlBase)
	if err != nil {
		return ""
	}
	u.Path = path.Join(u.Path, s.bucket, key)
	return u.String()
}
