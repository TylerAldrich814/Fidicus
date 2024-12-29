package config

import (
	"log"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv - Loads environment variables from local .env file if present.
func LoadEnv(){
  rootPath, err := filepath.Abs(".env")
  log.Printf("RootPath: %s", rootPath)
  if err != nil {
    log.Fatalf("Failed to resolve .env file path: %v\n", err)
  }

  err = godotenv.Load(rootPath)
  if err != nil {
    log.Println("No .env file found. Falling back to system environment variables.")
  }
}
