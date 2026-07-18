package storage

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// isNotFound reports whether err is an S3 not-found error.
func isNotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	var notFound *types.NotFound
	return errors.As(err, &noSuchKey) || errors.As(err, &notFound)
}
