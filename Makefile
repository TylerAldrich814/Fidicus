MKFILE_DIR  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

##  Postgres Development Commands
.PHONY: dev_pg_migrate dev_pg_up dev_pg_down dev_pg_remove dev_pg_reset dev_pg

dev_pg_migrate:## dev_pg_migrate :: calles go run cmd/migrate/main.go up for initializing our base auth migrations. This will be removed in the future.
	@go run cmd/migrate/main.go up


dev_pg_up:     ## dev_pg_up      :: Runs docker-compose up -d -- Creating a Postgres Docker Image.
	@docker-compose up -d

dev_pg_down:   ## dev_pg_down    :: Runs docker-compose down -- Stopping our Postgres Docker Image.
	@docker-compose down

dev_pg_remove: ## dev_pg_remove  :: Removes out Postgres Docker Image
	@docker volume rm schematix_postgres_data

dev_pg_reset:  ## dev_pg_reset   :: Shuts down our Postgres Docker image, then removes the image and finally rebuilds and runs our Postgres Image.
	@$(MAKE) dev_pg_down dev_pg_remove dev_pg_up 

dev_pg:        ## dev_pg         :: Runs Postgres CLI for Schemataix: arg migrate calls dev_pg_migrate
	@psql -h localhost -U admin -d schematix_auth

help:
	@echo "Available Commands:"
	@echo $(MAKEFILE_LIST)
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf " - %0s\n", $$2}'
