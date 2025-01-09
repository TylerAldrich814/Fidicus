package pgsql

import (
	"context"
	"time"

	// "github.com/jackc/pgx/v5/pgconn"
  // "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"

  "github.com/TylerAldrich814/Fidicus/internal/shared/utils"
	"github.com/TylerAldrich814/Fidicus/internal/shared/role"
	repo "github.com/TylerAldrich814/Fidicus/internal/schema/infrastructure/repository"
)

// SchemaPGSQL -- A Postgres wrapper that implements our Schema Metadat Repository
type SchemaPGSQL struct {
  db *pgxpool.Pool
}

// NewSchemaMetadataRepo - Creates a new Schema Metadata SQL Database connection, and returns
// a newly created SchemaPGSQL instance when successul.
func NewSchemaMetadataRepo(
  ctx context.Context,
  dsn string,
)( *SchemaPGSQL, error ) {
  var pushLog = utils.NewLogHandlerFunc(
    "NewSchemaMetadataRepo",
    log.Fields{
      "DSN": dsn,
    },
  )

  config, err := pgxpool.ParseConfig(dsn)
  if err != nil {
    pushLog(
      utils.LogErro,
      "failed to parse postgres database config: %s",
      err.Error(),
    )
    return nil, repo.ErrPGSQLConfigFailed
  }

  config.MaxConns = 10 
  config.MinConns = 1
  config.MaxConnIdleTime = 5 * time.Minute
  config.MaxConnLifetime = 1 * time.Hour

  pool, err := pgxpool.NewWithConfig(ctx, config)
  if err != nil {
    pushLog(
      utils.LogErro,
      "failed to create postgres pool with config: %s",
      err.Error(),
    )
    return nil, repo.ErrDBFailedCreation
  }

  if err := pool.Ping(ctx); err != nil {
    pushLog(
      utils.LogErro,
      "failed to ping newly created postgres pool: %s",
      err.Error(),
    )
    return nil, repo.ErrDBFailedPing
  }

  return &SchemaPGSQL{ pool }, nil
}

func(s *SchemaPGSQL) CreateAccessRole(
  ctx context.Context, 
  role role.Role,
) error {
  // CREATE ROLE entity_admin WITH LOGIN PASSWORD 'password';
  // GRANT CONNECT ON DATABASE schemas_db TO entity_admin;
  // GRANT USAGE ON SCHEMA public TO entity_admin;
  // GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO entity_admin;


  return nil
}
