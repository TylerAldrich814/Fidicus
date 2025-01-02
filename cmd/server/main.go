package main

import (
	"context"
	"os"
	"os/signal"
  "github.com/gorilla/mux"

	AuthService "github.com/TylerAldrich814/Fidicus/internal/auth/application"
	AuthRepo "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/repository"
	AuthHTTP "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/http"
	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
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

  r := mux.NewRouter()

  if err := authHTTPHandler.RegisterRoutes(r)  ; err != nil {
    panic(err)
  }
  
}
