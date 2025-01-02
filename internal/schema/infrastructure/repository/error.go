package repository

import "errors"

type BlobErr error
type GraphErr error

var (
  ErrBlobDBInternal             BlobErr = errors.New("an internal error occurred")
  ErrBlobDBMinioInit            BlobErr = errors.New("failed to initialize minio client")
  ErrBlobDBFailedToCreateBucket BlobErr = errors.New("failed to create bucket")
  ErrBlobDBUploadFailed         BlobErr = errors.New("failed to upload new file into Blob Storage")
)
