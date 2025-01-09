package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"

	AuthService "github.com/TylerAldrich814/Fidicus/internal/auth/application"
	AuthHTTP "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/http"
	AuthRepo "github.com/TylerAldrich814/Fidicus/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
	log "github.com/sirupsen/logrus"
)

var (
  PostgresAuth = config.GetEnv("DB_HOST", "5432")
  DevServer    = "localhost:8080"
)

func main(){
  // ->> App Config
  ctx, cancel := signal.NotifyContext(
    context.Background(),
    os.Interrupt,
  )
  defer cancel()

  config.InitLogger()

  // <TODO> Tracker: jaegar

  // ->> Auth Repository Initialization:
  dbConfig := config.GetPGSQLConfig()
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
    panic(fmt.Sprintf("failed to register auth routes: %s", err.Error()))
  }

  errMsg := make(chan error, 1)
  go func(){
    log.Info("STARTING SERVER")
    if err := http.ListenAndServe(DevServer, r); err != nil {
      errMsg<-fmt.Errorf("fialed to start HTTP Server: %s", err.Error())
    }
  }()

  select {
  case err := <-errMsg:
      log.Fatal(err) // Log error and exit
  case <-ctx.Done():
      log.Info("SHUTTING DOWN")
      shutdownCtx := context.Background()
      shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
      defer cancel()

      if err := authHTTPHandler.Shutdown(); err != nil {
          log.Error("Tried to shutdown auth HTTP handler. Already closed.")
      }
  }
}
