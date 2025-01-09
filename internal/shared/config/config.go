package config

import (
	"fmt"
)

// AppConfig - Defines Global Variable Parameters
type AppConfig struct {
  ServerPort  string
  Environment string
  LogLevel    string
}

// PGSQLConfig - Defines Postgres Database Configuration
type PGSQLConfig struct {
  Host      string
  Port      string
  User      string
  Passw     string
  DBName    string
  SSLMode   string
}

func GetAppConfig() AppConfig {
  return AppConfig{
    ServerPort  : GetEnv("SERVER_PORT", ""),
    Environment : GetEnv("ENVIRONMENT", ""),
    LogLevel    : GetEnv("LOG_LEVEL",   ""),
  }
}

func GetPgsqlConfig(service string)( PGSQLConfig, error ){
  dbName := ""
  switch service {
  case "-auth":
    dbName = "AUTH_PGSQL"
  case "-schema":
    dbName = "SCHEMA_PGSQL"
  default:
    return PGSQLConfig{}, fmt.Errorf("Unknown service name: \"%s\"", service)
  }

  return PGSQLConfig{
    DBName  : GetEnv(dbName,    ""),
    Host    : GetEnv(dbName + "_HOST",    ""),
    Port    : GetEnv(dbName + "_PORT",    ""),
    User    : GetEnv(dbName + "_USER",    ""),
    Passw   : GetEnv(dbName + "_PASSW",   ""),
    SSLMode : GetEnv(dbName + "_SSLMODE", ""),
  }, nil
}

// GetPostgresURI - Creates a Postgres URI string via PGSQLConfig variables
//     ->> postgres://USER:PASSW@HOST:PORT/DBNAME?sslmode=SSLMODE
func(d *PGSQLConfig) GetPostgresURI() string {
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

// var (
//   envInstance *env
//   once        sync.Once
// )
//
// type env struct { }
//
// func GetEnv(key, defaultValue string) string {
//   // if envInstance == nil {
//   // }
//   once.Do(func(){
//     godotenv.Load()
//   })
//
//   value := os.Getenv(key)
//   if value == "" {
//     return defaultValue
//   }
//   return value
// }
