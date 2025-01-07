package proto

import (
	"fmt"

	"github.com/bufbuild/protocompile/linker"
)

type ProtoMetadata struct{
  File    linker.File
  Imports []string
  PkgName string
  Path    string
}

// ProtoDependencyGraph -- Defines a Proto File Dependency graph.
// Which will determine in what order we comile our Cypher queries.
// We need to compile our queries from least specificity to greatest
// specificity in order to make sure all cypher relationship queries
// are successfully processed by Neo4j
type ProtoDependencyGraph struct {
  Ordered  []*ProtoMetadata          // Final Sorted Order
  packages map[string][]string       // File ->> Dependencies
  visited  map[string]bool           // Track visited Files.
  stack    map[string]*ProtoMetadata // Processsed Proto Metadat objects.
}

// BuildDependencyGraph - Builds a Dependency Graph Builder containing ProtoMetadata references,
// created for ever file passed within 'files'.
func BuildDependencyGraph(
  files linker.Files,
) ProtoDependencyGraph {
  graph := ProtoDependencyGraph {
    packages : map[string][]string{},
    visited  : map[string]bool{},
    stack    : map[string]*ProtoMetadata{},
    Ordered  : []*ProtoMetadata{},
  }
  var (
    pkgName    string
    imports    []string
    path string
  )

  for _, file := range files {
    pkgName  = string(file.Package())
    path     = string(file.Path())
    imports  = []string{}

    for i := 0; i < file.Imports().Len(); i++ {
      importPath := string(file.Imports().Get(i).Package())
      imports = append(imports, importPath)
    }
    graph.packages[pkgName] = imports
    graph.stack[pkgName] = &ProtoMetadata{
      File    : file,
      Imports : imports,
      PkgName : pkgName,
      Path    : path,
    }
  }

  return graph
}

// TopologicalSort - Sorts all Protofiles level of specificity.
// ProtoMetadata objects are sorted, in memory, within p.Ordered
func(p *ProtoDependencyGraph) TopologicalSort() error {
  var dfs func(node string)error

  dfs = func(node string)error {
    if p.visited[node] {
      return nil
    }
    p.visited[node] = true

    // ->> Visit All Dependencies recursively::
    for _, dep := range p.packages[node] {
      if err := dfs(dep); err != nil {
        return err
      }
    }

    meta, exists := p.stack[node]
    if !exists {
      return fmt.Errorf("metadata for %s not found", node)
    }
    p.Ordered = append(p.Ordered, meta)
    return nil
  }

  // ->> Perform DFS For Each Node:
  for node := range p.packages {
    if !p.visited[node] {
      // Detected Cycle:
      if err := dfs(node); err != nil {
        return err
      }
    }
  }

  return nil
}
