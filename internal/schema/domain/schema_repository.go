package domain

import (
	"context"
	"time"

	role "github.com/TylerAldrich814/Fidicus/internal/shared/domain"
)

// SchemaRepository defines our Schema's Storage Logic.
type SchemaBlobRepository interface {
  // UploadSchema -- Uploads a new Schema into Blob Database..
  UploadSchema(ctx context.Context, name, path string) error
  // DownloadSchema - Attempts to download a stored schema from DB..
  DownloadSchema(ctx context.Context, name string )([]byte, error)
  // DeleteSchema - Attempts to delete a Schema from blob DB.
  DeleteSchema(ctx context.Context, key string) error
  // ListSchemas -- Using a prefix, search and return an array of Object keys.
  ListSchemas(ctx context.Context, prefix string)([]string, error)
  // GeneratePresignedURL -- Generates pre-signed URLs for read/write access.
  GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration)( string, error)

  Shutdown()error
}

type SchemaGraphRepository interface {

  Shutdown()error
}

// SchemaSQLRepository defines our Schema's SQL Logic for storing and handling Schema Metaata.
type SchemaSQLRepository interface {
  CreateAccessRole(ctx context.Context, role role.Role) error

  Shutdown()error
}
