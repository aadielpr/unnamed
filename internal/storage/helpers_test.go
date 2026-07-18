package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aadielpr/unnamed/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
)

func createBucket(t *testing.T, cfg storage.Config) {
	t.Helper()

	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithBaseEndpoint(cfg.Endpoint),
	)
	require.NoError(t, err)

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	_, err = client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		var alreadyOwned *types.BucketAlreadyOwnedByYou
		var alreadyExists *types.BucketAlreadyExists
		if errors.As(err, &alreadyOwned) || errors.As(err, &alreadyExists) {
			return
		}
		require.NoError(t, err)
	}
}
