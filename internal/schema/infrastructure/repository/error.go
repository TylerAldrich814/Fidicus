package repository

import "errors"

type BlobErr  error
type PgSQLErr error
type GraphErr error

var (
  ErrBlobDBInternal             BlobErr = errors.New("an internal error occurred")
  ErrBlobDBMinioInit            BlobErr = errors.New("failed to initialize minio client")
  ErrBlobDBFailedToCreateBucket BlobErr = errors.New("failed to create bucket")
  ErrBlobDBUploadFailed         BlobErr = errors.New("failed to upload new file into blob storage")
  ErrBlobDBDownloadFailed       BlobErr = errors.New("failed to download file from blob storage")
  ErrBlobDBDeleteFailed         BlobErr = errors.New("failed to delete file from blob storage")
  ErrBlobDBIOError              BlobErr = errors.New("failed to read downloaded file from blob storage")

  ErrPGSQLConfigFailed            PgSQLErr = errors.New("failed to parse schema metadata DB config")
  ErrDBFailedCreation          PgSQLErr = errors.New("failed to create sql db pool")
  ErrDBFailedPing              PgSQLErr = errors.New("failed to ping newly created sql db")
)
