package config

import (
	"fmt"
	"os"
)

// AppConfig - Defines Global Variable Parameters
type AppConfig struct {
  ServerPort  string
  Environment string
  LogLevel    string
}

// DBConfig - Defines Postgres Database Configuration
type DBConfig struct {
  Host      string
  Port      string
  User      string
  Passw     string
  DBName    string
  SSLMode   string
}

func GetAppConfig() AppConfig {
  return AppConfig{
    ServerPort  : GetEnv("SERVER_PORT", "50051"),
    Environment : GetEnv("ENVIRONMENT", "development"),
    LogLevel    : GetEnv("LOG_LEVEL",   "debug"),
  }
}

func GetDBConfig() DBConfig {
  return DBConfig{
    Host    : GetEnv("DB_HOST",    "localhost"),
    Port    : GetEnv("DB_PORT",    "5432"),
    User    : GetEnv("DB_USER",    "admin"),
    Passw   : GetEnv("DB_PASSW",   "AdminPassword"),
    DBName  : GetEnv("DB_NAME",    "fidicus_auth"),
    SSLMode : GetEnv("DB_SSLMODE", "disable"),
  }
}

// GetPostgresURI - Creates a Postgres URI string via DBConfig variables
//     ->> postgres://USER:PASSW@HOST:PORT/DBNAME?sslmode=SSLMODE
func(d *DBConfig) GetPostgresURI() string {
  return fmt.Sprintf(
    "postgres://%s:%s@%s:%s/%s?sslmode=%s",
    d.User,
    d.Passw,
    d.Host,
    d.Port,
    d.DBName,
    d.SSLMode,
  )
}

func GetEnv(key, defaultValue string) string {
  value, exists := os.LookupEnv(key)
  if !exists {
    return defaultValue
  }
  return value
}
