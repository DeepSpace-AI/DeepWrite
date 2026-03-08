package storages

import "errors"

var (
	ErrUnsupportedProvider = errors.New("unsupported storage provider")
	ErrBucketRequired      = errors.New("storage bucket is required")
	ErrEndpointRequired    = errors.New("storage endpoint is required for this provider")
	ErrCredentialsRequired = errors.New("storage credentials are required")
	ErrObjectKeyRequired   = errors.New("object key is required")
)
