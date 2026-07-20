package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"
)

// createBucket ensures the test bucket exists, reusing the same client wiring
// as NewS3Store (newClient) instead of rebuilding the AWS config by hand.
func createBucket(t *testing.T, cfg Config) {
	t.Helper()

	client, err := newClient(cfg)
	require.NoError(t, err)

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
