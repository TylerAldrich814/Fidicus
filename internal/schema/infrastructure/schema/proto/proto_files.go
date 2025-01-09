package proto

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bufbuild/protocompile"
	"github.com/bufbuild/protocompile/linker"

	log "github.com/sirupsen/logrus"
)

type ProtoFiles struct {
  filepaths []string
  files     []linker.File
  parsed    map[string]*ProtoCypherCompiler
  compiled  map[string]*ProtoCypherCompiler
  depGraph  ProtoDependencyGraph
}

// NewLocal --- For Testing on local proto files.
func NewLocalFiles(
  ctx          context.Context,
  src          string,
  filepaths    []string,
)( *ProtoFiles, error ){
  // <NOTE> For Local Files
  sources := loadProtoSources(src)
  compiler := protocompile.Compiler{
    Resolver: &protocompile.SourceResolver{
      Accessor: protocompile.SourceAccessorFromMap(sources), 
    },
  }

  files, err := compiler.Compile(ctx, filepaths...)
  if err != nil {
    log.Error("Failed to compile proto files: %s", err.Error())
    return nil, err
  }

  depGraph := BuildDependencyGraph(files)
  depGraph.TopologicalSort()

  return &ProtoFiles{
    filepaths : filepaths,
    files     : files,
    parsed    : make(map[string]*ProtoCypherCompiler),
    depGraph  : depGraph,
  }, nil
}

// NewProtoFiles - creates a ProtoFiles instance. Settings up both
// protocompile and our Proto Dependency Graph.
func NewBlobFiles(
  ctx          context.Context,
  filepaths    []string,
  protoHandler *ProtoHandler,
)( *ProtoFiles, error ){
  // For BlobDB Stored Files
  compiler := protocompile.Compiler{
    Resolver: &protocompile.SourceResolver{
      Accessor: protoHandler.GenerateResolver(ctx),
    },
  }

  files, err := compiler.Compile(ctx, filepaths...)
  if err != nil {
    log.Error("Failed to compile proto files: %s", err.Error())
    return nil, err
  }

  depGraph := BuildDependencyGraph(files)
  depGraph.TopologicalSort()

  return &ProtoFiles{
    filepaths : filepaths,
    files     : files,
    parsed    : make(map[string]*ProtoCypherCompiler),
    depGraph  : depGraph,
  }, nil
}

func loadProtoSources(root string) map[string]string {
  sources := make(map[string]string)
  _ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
    if err == nil && strings.HasSuffix(path, ".proto") {
      data, readErr := os.ReadFile(path)
      if readErr == nil {
        relPath, _ := filepath.Rel(root, path)
        sources[relPath] = string(data)
      }
    }
    return nil
  })

  return sources
}

func(pf *ProtoFiles) ParseSchemaFiles(
  ctx context.Context,
) error {
  // Create Dependency Graph to determine which order to parse schemas. 

  compilers := make(map[string]*ProtoCypherCompiler)
  errors := make(map[string]error)
  var wg sync.WaitGroup

  for _, dep := range pf.depGraph.Ordered {
    fmt.Printf("Processing Package: \"%s\"\n",dep.PkgName)
    fmt.Printf("Dependencies: \"%s\"\n", strings.Join(dep.Imports, ", "))

    file := dep.File
    if file.Syntax().String() != "proto3" {
      return fmt.Errorf("currently only support proto3: not %s", file.Syntax().String())
    }
    compilers[dep.PkgName] = NewProtoCypherCompiler()

    wg.Add(1)
    go func(key string, protoMetadata *ProtoMetadata, errors map[string]error){
      defer wg.Done()
      if err := compilers[key].Run(protoMetadata); err != nil {
        errors[protoMetadata.PkgName] = fmt.Errorf("")
      }
    }(dep.PkgName, dep, errors)
  }


  wg.Wait()
  time.Sleep(time.Second)

  // <TODO> :: This is temporary. Will need to create an error handling message broker
  err := new(strings.Builder)
  for p, e := range errors {
    if e != nil {
      err.WriteString(fmt.Sprintf(
        "Failed to compile \"%s\": %s\n",
        p,
        e.Error(),
      ))
    }
  }
  errMsg := err.String()
  if errMsg != "" {
    return fmt.Errorf(errMsg)
  }

  pf.compiled = compilers
  return nil
}

// Cypher -- Combines all of the parsed Cypher Queries from the provided files
// in order from least specificity to greatest specificity.
func(pf *ProtoFiles) Cyphers()( []string, error ){
  if pf.compiled == nil || len(pf.compiled) == 0 {
    return nil, fmt.Errorf("Files have not been compiled yet")
  }
  cyphers := []string{}

  for _, dep := range pf.depGraph.Ordered {
    cyphers = append(
      cyphers, 
      strings.TrimSpace(pf.compiled[dep.PkgName].WriteString()),
    )
  }
  return cyphers, nil
}
