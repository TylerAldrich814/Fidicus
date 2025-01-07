package tests

import (
	"bufio"
	"context"
	"os"
	"testing"

	"github.com/TylerAldrich814/Fidicus/internal/schema/infrastructure/schema"
	// "github.com/stretchr/testify/assert"
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

// TestSchemaParser - Pulls in both test.proto and shouldbe.cypher paths.
// parsers test.proto through Proto.ParseSchemaFile by pasing in a Test 
// callback funciton instead of a GraphDB Query Runner.
func TestSchemaParser(t *testing.T){
  testProtoFile := "test.proto"

  // expectedCypherFile := "shouldbe.cypher"
  // expectedQueries, err := readLines(expectedCypherFile)
  // if err != nil {
  //   t.Errorf("ReadLines Failure: %s", err.Error())
  // }
  proto := schema.NewProto()
  testParsedSchema := func(queries string) error {
    // We need to condense Parsed Queries into a single string and then split it per line.
    // condensed := strings.Split(strings.Join(queries, ""), "\n")
    //
    // for _, q := range condensed {
    //   t.Log(strings.TrimSpace(q))
    // }
    // t.Error()

    // for i, q := range condensed {
    //   got := strings.TrimSpace(q)
    //   want := strings.TrimSpace(expectedQueries[i])
    //
    //   assert.Equal(t, got, want, "Query Line must be equvalent")
    // }
    return nil
  }

  if err := proto.ParseSchemaFile(
    context.Background(), 
    testProtoFile, 
    testParsedSchema,
  ); err != nil {
    t.Fatalf("failed to parse proto file: %v", err)
  }
}
