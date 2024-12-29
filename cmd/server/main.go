package main

import (
	"github.com/TylerAldrich814/Schematix/internal/auth/application"
	"github.com/TylerAldrich814/Schematix/internal/auth/infrastructure/repository"
	"github.com/TylerAldrich814/Schematix/pkg/jwt"
)

func main(){
  repo := repository.New()
  jwt  := jwt.New()
  s := application.NewService(
    repo,
    jwt,
  )
}


