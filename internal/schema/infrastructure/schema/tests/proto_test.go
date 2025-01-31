package tests

import (
	"bufio"
	"os"
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
