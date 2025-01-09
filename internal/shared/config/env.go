package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
  once sync.Once
  envinstance *envLoaded
)
type envLoaded struct {}

func GetEnv(key, defaultVal string) string {
  if envinstance == nil {
    once.Do(loadEnv)
    envinstance = &envLoaded{}
  }
  e := os.Getenv(key)
  if e == "" { 
    return defaultVal
  }
  return e
}

// loadEnv - Loads environment variables from Root .env.
func loadEnv(){
  wd, err := os.Getwd()
  if err != nil {
    log.Error(fmt.Sprintf("laodEnv: Failed to retreive current Working Directory: %s", err))
  }
  parts := strings.Split(wd, "Fidicus")
  if len(parts) == 0 {
    log.Error(fmt.Sprintf("loadEnv: Not in a Fidicus sub directory"))
    return
  }
  rootDir := parts[0] + "Fidicus/.env"
  log.Printf("loadEnv: RootDirectory: %s", rootDir)

  if err := godotenv.Load(rootDir); err != nil {
    log.Println("No .env file found. Falling back to system environment variables.")
  }
}
