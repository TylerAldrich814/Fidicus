package domain

import "context"

// Schema defined a Microservice Messaging Protocol interface.
// i.e., gRPC, GraphQL, OpenAPI, etc.
type Schema interface {
  // ParseSchemaFile(ctx context.Context, filename string, runner func(queries []string)) error
  ParseSchemaFiles(ctx context.Context) error
  // Cyphers Provides an array of compiled cypher queries.
  Cyphers()( []string, error )
}
