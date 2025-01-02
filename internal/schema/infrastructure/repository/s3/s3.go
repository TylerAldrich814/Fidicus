package s3

import (
	"context"

	"github.com/TylerAldrich814/Fidicus/internal/schema/infrastructure/repository"
	"github.com/TylerAldrich814/Fidicus/internal/shared/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	log "github.com/sirupsen/logrus"
)

// S3Storage - defines our Fidicus
type S3Storage struct {
  client *minio.Client
  bucket string
}

// NewS3Storage - Creates a new S3Storage instance.
//
// Potential Erros:
//    - BlobErr.ErrBlobDBInternal
//    - BlobErr.ErrBlobDBMinioInit
//    - BlobErr.ErrBlobDBFailedToCreateBucket
func NewS3Storage(
  ctx       context.Context,
  endpoint  string,
  accessKey string,
  secretKey string,
  bucket    string,
  useSSL    bool,
)( *S3Storage, error ){
  var pushLog = utils.NewLogHandlerFunc(
    "NewS3Storage",
    log.Fields{
      "endpoint"  : endpoint,
      "accessKey" : accessKey,
      "secretKey" : secretKey,
      "bucket"    : bucket,
      "useSSL"    : useSSL,
    },
  )

  client, err := minio.New(
    endpoint,
    &minio.Options{
      Creds: credentials.NewStaticV4(
        accessKey,
        secretKey,
        "",
      ),
    },
  )
  if err != nil {
    pushLog(utils.LogErro, "failed to create new minio client: %s", err.Error())
    return nil, repository.ErrBlobDBMinioInit
  }

  exists, err := client.BucketExists(
    ctx,
    bucket,
  )
  if err != nil {
    pushLog(utils.LogErro, "failed to check if bucket exists: %s", err.Error())
    return nil, repository.ErrBlobDBInternal
  }
  if !exists {
    err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
    if err != nil {
      pushLog(utils.LogErro, "failed to create new bucket: %s", err.Error())
      return nil, repository.ErrBlobDBFailedToCreateBucket
    }
  }

  return &S3Storage{
    client, 
    bucket,
  }, nil
}
 
// UploadSchema -- Uploads a new Schema into our S3 Bucket.
func(s *S3Storage) UploadSchema(
  ctx context.Context, 
  name string, 
  path string,
) error {
  var pushLog = utils.NewLogHandlerFunc(
    "UploadSchema",
    log.Fields{
      "name" : name,
      "path" : path,
    },
  )

  if _, err := s.client.FPutObject(
    ctx,
    s.bucket,
    name,
    path,
    minio.PutObjectOptions{},
  ); err != nil {
    pushLog(utils.LogErro, "failed to put new object into storage: %s", err.Error())
    return repository.ErrBlobDBUploadFailed
  }

  return nil
}
 
 
 
 
 
 
 
 
 
 
 
 
 
 
 
 
 
 
