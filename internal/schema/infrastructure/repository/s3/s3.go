package s3

import (
	"context"
	"io"
	"net/url"
	"time"

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
 
// DownloadSchema - Attempts to download a stored schema from s3.
func(s *S3Storage) DownloadSchema(
  ctx context.Context,
  name string,
)( []byte, error ){
  var pushLog = utils.NewLogHandlerFunc(
    "DownloadSchema",
    log.Fields{
      "name": name,
    },
  )

  obj, err := s.client.GetObject(
    ctx,
    s.bucket,
    name,
    minio.GetObjectOptions{},
  )
  if err != nil {
    pushLog(
      utils.LogErro,
      "failed to get stored schema from s3: %s",
      err.Error(),
    )
    return nil, repository.ErrBlobDBDownloadFailed
  }
  defer obj.Close()

  schema, err := io.ReadAll(obj)
  if err != nil {
    pushLog(
      utils.LogErro,
      "failed to read received schema from s3: %s",
      err.Error(),
    )
    return nil, repository.ErrBlobDBIOError
  }

  return schema, nil
}

// DeleteSchema - Attempts to delete a schema from S3.
func(s *S3Storage) DeleteSchema(
  ctx context.Context,
  key string,
) error {
  if err := s.client.RemoveObject(
    ctx,
    s.bucket,
    key,
    minio.RemoveObjectOptions{},
  ); err != nil {
    utils.NewLogHandlerFunc(
      "DeleteSchema",
      log.Fields{ "key": key },
    )(utils.LogErro, "failed to remove schema from s3: %s", err.Error())
    return repository.ErrBlobDBDeleteFailed
  }

  return nil
}

  // GeneratePresignedURL -- Generates pre-signed URLs for read/write access.
func(s *S3Storage) GeneratePresignedURL(
  ctx context.Context, 
  key string, 
  expiry time.Duration,
)( string, error) {
  reqParams := make(url.Values)
  presignedURL, err := s.client.PresignedGetObject(
    ctx,
    s.bucket,
    key,
    expiry,
    reqParams,
  )
  if err != nil {
    return "", nil
  }

  return presignedURL.String(), nil
}
 
// ListSchemas -- Using a prefix, search and return an array of 
// Object keys.
func(s *S3Storage) ListSchemas(
  ctx    context.Context,
  prefix string,
)( []string, error ){
  var pushLog = utils.NewLogHandlerFunc(
    "DownloadSchema",
    log.Fields{
      "prefix": prefix,
    },
  )

  objectCh := s.client.ListObjects(
    ctx,
    s.bucket,
    minio.ListObjectsOptions{
      Prefix: prefix,
      Recursive: true,
    },
  )

  var schemas []string
  for object := range objectCh {
    if object.Err != nil {
      pushLog(
        utils.LogErro,
        "listed schema resulted in an error: %s",
        object.Err.Error(),
      )
      return nil, repository.ErrBlobDBInternal
    }
    schemas = append(schemas, object.Key)
  }

  return schemas, nil
}
 
 
 
 
 
 
 
 
 
 
 
 
 
 
