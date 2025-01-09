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
