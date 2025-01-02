package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main(){
  config.LoadEnv()

  // Logger Setup:
  config.InitLogger()

  // Database Configs:
  dbConfig  := config.GetDBConfig()

  // Postgres Connextion:
  dsn := dbConfig.GetPostgresURI()
  log.Printf("DSN \"%s\"", dsn)

  db, err := sql.Open("postgres", dsn)
  if err != nil {
    log.Fatalf("Failed to connect to postgres: %v\n", err)
  }
  defer db.Close()

  // Create Postgres Migration Driver:
  driver, err := postgres.WithInstance(db, &postgres.Config{})
  if err != nil {
    log.Fatalf("Failed to create migration driver: %v\n", err)
  }

  // Create Migration Instance
  m, err := migrate.NewWithDatabaseInstance(
    "file://cmd/migrate/migrations",
    "postgres",
    driver,
  )
  if err != nil {
    log.Fatalf("Failed to initialize postgres migration: %v\n", err)
  }

  // Handle CLI Commands
  if len(os.Args) < 2 {
    log.Fatalf("Usage: go run main.go [up|down|force VERSION|version]")
  }

  switch os.Args[1] {
  case "up":
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
      log.Fatalf("Failed to apply migrations: %v\n", err)
    }
    log.Println("Migrations applied Successfully.")
  case "down":
    if err := m.Down(); err != nil && err != migrate.ErrNoChange {
      log.Fatalf("Failed to rollback migrations: %v\n", err)
    }
    log.Println("Migrations rolled back successfully.")
  case "force":
    if len(os.Args) < 3 {
      log.Fatalf("Usage: go run main.go force VERSION")
    }
    version, err := strconv.Atoi(os.Args[2])
    if err != nil {
      log.Fatalf("Invalid version number: %v\n", err)
    }
    if err := m.Force(version); err != nil {
      log.Fatalf("Failed to force migration version: %v\n", err)
    }
    log.Printf("Migration force to version: %d\n", version)
  case "version":
    version, dirty, err := m.Version()
    if err != nil {
      log.Fatalf("Failed to get migration version: %v", err)
    }
    log.Printf("Current migration version: %d, Dirty: %v", version, dirty)

  default:
    log.Fatalf("Unknown command: %s\n", os.Args)
  }
}
