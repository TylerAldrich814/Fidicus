include .env
export $(shell sed 's/=.*//' .env)

MKFILE_DIR  := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

##  Postgres Development Commands
.PHONY: go_replace dev_mgrok dev_auth_migrate dev_docker_up dev_docker_down dev_pg_remove dev_neo_remove dev_pg_reset dev_pg

go_replace:      ## go_replace       :: Requires two cli args: from="SomeString" to="ToString" -- Taking both arguments, we recursively finds and replace all 'to's with 'from's in all *.go files under ./Fidicus
	@if [ "$$from" = "" ]; then             \
		echo "Missing 'from' input argument"; \
		exit 1;                               \
	fi;
	@if [ "$$to" = "" ]; then               \
		echo "Missing 'to' input argument";   \
		exit 1;                               \
	fi;
	@find . -type f -name "*.go" -exec sed -i '' 's/$$from/$$to/g' {} +

dev_mgrok:       ## dev_mgrok       :: Starts up our mgrok server -- allowing us to test our development application online
	@ngrok http --url=on-shad-capable.ngrok-free.app 8080

dev_docker_up:   ## dev_docker_up   :: Runs docker-compose up -d -- Creating a Postgres Docker Image.
	@docker-compose up -d

dev_docker_down: ## dev_docker_down :: Runs docker-compose down -- Stopping our Postgres Docker Image.
	@docker-compose down

dev_auth_migrate:  ## dev_pg_migrate  :: calles go run cmd/migrate/main.go up for initializing our base auth migrations. This will be removed in the future.
	@go run cmd/migrate/main.go -pg -auth up

dev_pg_remove:   ## dev_pg_remove   :: Removes out Postgres Docker Image
	@docker volume rm fidicus_postgres_data

dev_neo_remove:  ## dev_neo_remove  :: Removes all Fidicus Neo4J Volumes from docker.
	@docker volume rm $(docker volume ls -q | grep "fidicus_neo4j")

dev_pg_reset:    ## dev_pg_reset    :: Shuts down our Postgres Docker image, then removes the image and finally rebuilds and runs our Postgres Image.
	@$(MAKE) dev_docker_down dev_pg_remove dev_docker_up 

dev_pg:          ## dev_pg          :: Runs Postgres CLI for Schemataix: arg migrate calls dev_pg_migrate
	@psql -h localhost -U admin -d fidicus_auth

dev_neo4j:
	@docker exec -it fidicus-neo4j bin/cypher-shell -u ${NEO4J_USER} -p ${NEO4J_PASSWORD}

help:
	@echo "Available Commands:"
	@echo $(MAKEFILE_LIST)
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf " - %0s\n", $$2}'
