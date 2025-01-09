package main

import (
	"log"
	"os"
	"strings"

	"github.com/TylerAldrich814/Fidicus/cmd/migrate/db"
	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main(){
  config.InitLogger()

  if strings.ToLower(os.Args[1]) == "help" || os.Args[1] == "-h" {
    log.Println(" --- Fidicus MIgration Tool ---")
    log.Println("  -- Expects 3 sepeate commands:")
    log.Println("  go run main.go -DB_ID XX XX :: 3 possible values, tells Migrate which database type to target")
    log.Println("      | -pg")
    log.Println("      | -neo")
    log.Println("      | -s3")
    log.Println(" go run main.go -DB_ID -SERVICE XX :: 2 possible values, tells Migrate which service to target.")
    log.Println("      | -auth")
    log.Println("      | -schema")
    log.Println(" go run main.go -DB_ID -SERVICE ACTION :: 4 possible values, tells Migrate which action to use.")
    log.Println("      | up")
    log.Println("      | down")
    log.Println("      | force")
    log.Println("      | version")
    log.Println(" NOTE: Not all services use the same databased. Currently, schema is the only service that uses all 3 database options")
    log.Println("       Auth only uses Postgres. The purpose of this tool is for streamlining database actions between all services. As")
    log.Println("       this project grows, the more we'll need this script for updating our migrations.")
  }

  if len(os.Args) < 4 {
    log.Fatalf("Usage: go run main.go [-pg | -neo | -s3] [ -auth | -schema ] up | down | force [VERSION|version]")
  }
  // Setup Databases:
  database, svc, cmd := os.Args[1], os.Args[2], os.Args[3]

  if svc != "-auth" && svc != "-schema" {
    log.Fatalf("Unknown service: \"%s\"", svc)
  }
  if svc == "-auth" && database != "-pg" {
    log.Fatal("Auth service uses has Postgres, not \"%s\"", svc)
  }

  switch database{
  case "-pg":
    if err := db.PostgresMigrataion(svc, cmd); err != nil  {
      log.Fatal(err)
    }
  case "-neo":

  case "-s3":
  default:
    log.Fatalf("Unknown database: \"%s\"", os.Args[1])
  }
}

