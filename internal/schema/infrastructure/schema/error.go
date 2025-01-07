package schema

import "errors"

var (
  ErrSchemaParseFailed = errors.New("failed to parse schema file")
  ErrSchemaQueryRunner = errors.New("processing queries failed")
)
