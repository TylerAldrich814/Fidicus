package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	AuthService "github.com/TylerAldrich814/Schematix/internal/auth/application"
	AuthRepo "github.com/TylerAldrich814/Schematix/internal/auth/infrastructure/repository"
	AuthHTTP "github.com/TylerAldrich814/Schematix/internal/auth/infrastructure/http"
	"github.com/TylerAldrich814/Schematix/internal/shared/config"
)

var (
  PostgresAuth = config.GetEnv("DB_HOST", "5432")
)

func main(){
  // ->> App Config
  config.LoadEnv()

  ctx, cancel := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
  )
  defer cancel()

  config.InitLogger()

  // <TODO> Tracker: jaegar

  // ->> Auth Repository Initialization:
  log.Warn()
  dbConfig := config.GetDBConfig()
  dsn := dbConfig.GetPostgresURI()

  authRepo, err := AuthRepo.NewAuthRepo(
    ctx, 
    dsn,
  )
  if err != nil {
    panic("Failed to start Auth Repository")
  }
  authService := AuthService.NewService(
    authRepo,
  )
  authHTTPHandler := AuthHTTP.NewHttpHandler(authService)

  mux := http.NewServeMux()

  if err := authHTTPHandler.RegisterRoutes(mux)  ; err != nil {
    
    panic(err)
  }
  
}
