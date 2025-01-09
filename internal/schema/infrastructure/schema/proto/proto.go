package proto

import (
	"io"
	"strings"
  "context"

	"github.com/TylerAldrich814/Fidicus/internal/schema/domain"
)

// ProtoHandler defines a data structure used for Loading Blob Stored Proto files into 
// our Proto Cypher Compiler.
type ProtoHandler struct {
  registry domain.SchemaBlobRepository
}

func NewProtoHandler(
  registry domain.SchemaBlobRepository,
) *ProtoHandler {
  return &ProtoHandler{ registry }
}

// GenerateResolver - Defines a function factory for creating a File Resolver
// for protocompile to use to load in proto files from our BlobDB
func(p *ProtoHandler) GenerateResolver(
  ctx context.Context,
) func(string)(io.ReadCloser, error) {
  return func(path string)(io.ReadCloser, error) {
    obj, err := p.registry.DownloadSchema(ctx, path)
    if err != nil {
      return nil, err
    }
    return io.NopCloser(strings.NewReader(string(obj))), nil
  }
}
