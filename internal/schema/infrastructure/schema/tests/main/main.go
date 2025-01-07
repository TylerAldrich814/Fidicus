package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TylerAldrich814/Fidicus/internal/schema/infrastructure/schema/proto"
)

func readLines(filename string)( []string, error ){
  file, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  var lines []string
  scanner := bufio.NewScanner(file)
  for scanner.Scan(){
    lines = append(lines, scanner.Text())
  }

  return lines, scanner.Err()
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

// TestSchemaParser - Pulls in both test.proto and shouldbe.cypher paths.
// parsers test.proto through Proto.ParseSchemaFile by pasing in a Test 
// callback funciton instead of a GraphDB Query Runner.
func main(){
  start := time.Now()
  testProtoFiles := []string{"common.proto", "user.proto", "other.proto", "third.proto"}
  ctx := context.Background()

  protoFiles, err := proto.NewProtoFiles(ctx, testProtoFiles)
  if err != nil {
    panic(err)
  }

  if err := protoFiles.ParseSchemaFiles(ctx); err != nil {
    panic(err)
  }
  cyphers, err := protoFiles.Cyphers()
  if err != nil {
    panic(err)
  }

  for _, c := range cyphers {
    fmt.Println(c)
  }

  fmt.Printf(" ------- Total Time: %dms ------- \n", time.Since(start).Milliseconds())
}
