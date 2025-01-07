package domain

import "context"

// SchemaRepository defines our Schema's Storage Logic.
type SchemaBlobRepository interface {
  UploadSchema(ctx context.Context, name, path string) error
}

type SchemaGraphRepository interface {

}

// SchemaSQLRepository defines our Schema's SQL Logic for storing and handling Schema Metaata.
type SchemaSQLRepository interface {

}
