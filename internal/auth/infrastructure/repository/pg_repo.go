package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TylerAldrich814/Schematix/internal/auth/domain"
	"github.com/TylerAldrich814/Schematix/internal/shared/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

// PGRepo -- A Postgres wrapper that implements AuthRepository
type PGRepo struct {
  db *pgxpool.Pool
}

// New -- Creates a new PGRepo instance.
func New(
  ctx context.Context,
)( *PGRepo, error ){
  config.LoadEnv()

  dbConfig := config.GetDBConfig()
  dsn := dbConfig.GetPostgresURI()

  config, err := pgxpool.ParseConfig(dsn)
  if err != nil {
    log.Error("Failed to Parse Postgres Database Config: " + err.Error())
    return nil, ErrDBConfigParse
  }

  config.MaxConns = 10
  config.MinConns = 1
  config.MaxConnIdleTime = 5 * time.Minute
  config.MaxConnLifetime = 1 * time.Hour

  pool, err := pgxpool.NewWithConfig(ctx, config)
  if err != nil {
    log.Error("Failed to create Postgres Pool with Config: " + err.Error())
    return nil, ErrDBFailedCreation
  }

  if err := pool.Ping(ctx); err != nil {
    log.Error("Failed to Ping Database with newly created Postgres Pool: " + err.Error())
    return nil, ErrDBFailedPing
  }
  
  return &PGRepo{ pool }, nil
}

// Close -- Closes the Postgres Pool Connection.
func( pg *PGRepo ) Close(){
  pg.db.Close()
}

// CreateEntity -- Creates a new Schematix Entity with Root Privileges.
func(pb *PGRepo) CreateEntity(
  ctx    context.Context, 
  entity domain.Entity,
  account   domain.Account,
)( domain.EntityID, domain.AccountID, error) {
  // Generate IDs
  entityID := uuid.New()
  accountID   := uuid.New()
  entity.AccountIDs = append(entity.AccountIDs, accountID)

  var exists bool
  qErr := pb.db.QueryRow(
    ctx,
    `SELECT EXISTS(SELECT 1 FROM entities WHERE name = $1)`,
    entity.Name,
  ).Scan(&exists)

  if qErr != nil {
    return "", "", qErr
  }
  if exists {
    return "", "", ErrDBEntityAlreadyExists
  }

  qErr = pb.db.QueryRow(
    ctx,
    `SELECT EXISTS(SELECT 1 FROM accounts WHERE email = $1)`,
    account.Email,
  ).Scan(&exists)
  if qErr != nil {
    return "", "", qErr
  }
  if exists {
    return "", "", ErrDBAccountAlreadyExists
  }

  tx, err := pb.db.Begin(ctx)
  if err != nil {
    return "", "", err
  }
  defer tx.Rollback(ctx)

  var logError = func(e string){
    log.WithFields(log.Fields{
      "entity" : entity.Name,
      "account"   : account.Email,
    }).Error("CreateEntity: " + e)
  }

  _, err = tx.Exec(
    ctx,
    `INSERT INTO entities (
       id,
       name,
       description,
       account_ids,
       created_at,
       updated_at
     )
     VALUES ($1, $2, $3, $4, $5, $6)`,
    entityID, 
    entity.Name, 
    entity.Description, 
    entity.AccountIDs, 
    entity.CreatedAt, 
    entity.UpdatedAt,
  )
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
      logError("entity name already taken: " + err.Error())
      return "", "", ErrDBEntityAlreadyExists
    }
    logError("entity creation failed: " + err.Error())
    return  "", "", err
  }
  _, err = tx.Exec(
    ctx,
    `INSERT INTO accounts (
       id,
       entity_id,
       email,
       password_hash,
       role,
       first_name,
       last_name,
       cellphone_number,
       created_at,
       updated_at
     )
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
    accountID, 
    entityID,
    account.Email, 
    account.PasswHash,
    account.Role,
    account.FirstName,
    account.LastName,
    account.CellphoneNumber,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Error() == "23505" {
      logError("Account email already taken: " + err.Error())
      return "", "", ErrDBAccountAlreadyExists
    }
    logError("account creation failed: " + err.Error())
    return "", "", err
  }

  if err := tx.Commit(ctx); err != nil {
    logError("Failed to commit DB transaction: " + err.Error())
    return "", "", ErrDBFailedToCommitTX
  }

  return domain.EntityID(entityID.String()), domain.AccountID(accountID.String()), nil
}

// CreateAccount - Creates a new Entity Account.
func(p *PGRepo) CreateAccount(
  ctx context.Context, 
  eid  domain.EntityID,
  account domain.Account,
)( domain.AccountID, error){
  tx, err := p.db.Begin(ctx)
  if err != nil {
    return "", err
  }

  accountID := uuid.New()

  _, err = tx.Exec(
    ctx,
    `INSERT INTO accounts (
       id,
       entity_id,
       email,
       password_hash,
       role,
       first_name,
       last_name,
       cellphone_number,
       created_at,
       updated_at
     )
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
    accountID, 
    eid,
    account.Email, 
    account.PasswHash,
    account.Role,
    account.FirstName,
    account.LastName,
    account.CellphoneNumber,
    time.Now(),
    time.Now(),
  )
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
      return "", ErrDBAccountAlreadyExists
    }
    
    return  "", err
  }
  return domain.AccountID(accountID.String()), nil
}

func(pg *PGRepo) queryAndRemoveByID(
  ctx   context.Context,
  table string,
  key   string,
  id    string,
) error {
  var logError = func(e string) {
    log.WithFields(log.Fields{
      "table" : table,
      "id"    : id,
    }).Error("queryAndRemoveByID: " + e)
  }
  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("Failed to begin DB transaction: " + err.Error())
    return err
  }
  defer tx.Rollback(ctx)

  _, err = tx.Exec(
    ctx,
    `DELETE FROM $1 WHERE $2 = $3`,
    table,
    key,
    id,
  )
  if err != nil {
    logError("failed to delete row: " + err.Error())
    return err
  }

  err = tx.Commit(ctx)
  if err != nil {
    logError("failed to commit DB Transation: " + err.Error())
    return ErrDBFailedToCommitTX
  }

  return nil
}

// RemoveEntityByID - <TEMP> Removes an entity via it's ID
func(pg *PGRepo) RemoveEntityByID(
  ctx context.Context, 
  id string,
) error {
  var logError = func(e string) {
    log.WithFields(log.Fields{
      "id": id,
    }).Error("RemoveEntityByID: " + e)
  }

  // Begin TX
  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("failed to create new DB Transaction: " + err.Error())
    return ErrDBFailedToBeginTX
  }

  // Delete Entity
  if _, err := tx.Exec(
    ctx,
    `DELETE FROM entities WHERE id = $1`,
    id,
  ); err != nil {
    logError("failed to delete entity: " + err.Error())
    return ErrDBFailedToDeleteEntity
  }

  // Commit Changes
  if err := tx.Commit(ctx); err != nil {
    logError("failed to commit DB Transaction: " + err.Error())
    return ErrDBFailedToCommitTX
  }

  return nil
}

// RemoveEntity - <TEMP> Removes an entity by passing the name
func(pg *PGRepo) RemoveEntityByName(
  ctx        context.Context,
  entityName string,
) error {
  eid, err := pg.GetEntityIdByName(ctx, entityName)
  if err != nil {
    log.Error("RemoveEntityByName: Failed to query Entity ID: " + err.Error())
    return err
  }
  return pg.RemoveEntityByID(ctx, string(eid))
}



// RemoveAccountByID    - <TEMP> Removes an Account via it's ID
func(pg *PGRepo) RemoveAccountByID(
  ctx context.Context, 
  id string,
) error {
  var logError = func(e string) {
    log.WithFields(log.Fields{
      "id" : id,
    }).Error("RemoveAccountByID: " + e)
  }

  // Start new Transaction
  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("failed to begin new DB Transaction: " + err.Error())
    return ErrDBFailedToBeginTX
  }

  // Attempt to remove Account
  if _, err := tx.Exec(
    ctx,
    `DELETE FROM accounts WHERE id = $1`,
    id,
  ); err != nil {
    logError("Failed to delete account: " + err.Error())
    return ErrDBFailedToDeleteAccount
  }

  // Commit new Transaction
  if err := tx.Commit(ctx); err != nil {
    logError("Failed to commit DB Transaction: " + err.Error())
    return ErrDBFailedToCommitTX
  }

  return nil
}

// RemoveAccountByEmail - <TEMP> Removes an Account via it's Email
func(pg *PGRepo) RemoveAccountByEmail(
  ctx context.Context, 
  email string,
) error {
  id, err := pg.GetEntityIdByName(ctx, email)
  if err != nil {
    return err
  }

  return pg.RemoveAccountByID(ctx, string(id))
}

  // GetEntityIdByName - Query and returns an Entitys ID via it's name.
func(pg *PGRepo) GetEntityIdByName(
  ctx  context.Context, 
  name string,
)( domain.EntityID, error ){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "name": name,
    }).Error("GetEntityIdByName: " + e)
  }

  var entityID domain.EntityID

  err := pg.db.QueryRow(
    ctx,
    `SELECT id FROM entities WHERE name = $1`,
    name,
  ).Scan(&entityID)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("No Rows with name found")
      return "", ErrDBEntityNotFound
    }
    logError("Unknown error occurred: " + err.Error())
    return "", err
  }

  return entityID, nil
}

// GetAccountIdByName - Query and returns an Accounts ID via it's Email.
func(pg *PGRepo) GetAccountIdByName(
  ctx  context.Context, 
  email domain.AccountID,
)( domain.AccountID, error ){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "email": email,
    }).Error("GetAccountIdByName: " + e)
  }

  var accountID domain.AccountID

  if err := pg.db.QueryRow(
    ctx,
    `SELECT id FROM accounts WHERE email = $1`,
    email,
  ).Scan(&accountID); err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("no Rows with email found")
      return "", ErrDBAccountNotFound
    }
    logError("Unknown error occurred: " + err.Error())
    return "", err
  }

  return accountID, nil
}

// AccountSignin - Before signing account in. First detects if account exists then checks to make sure account is a Subaccount of Entity.
func(pg *PGRepo) AccountSignin(
  ctx   context.Context, 
  creds domain.Credentials,
)( *domain.AuthToken, error){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "Entity"    : creds.EntityName,
      "email"     : creds.Email,
      "pw_exists" : len(creds.Password) != 0,
    }).Error("AccountSignin: " + e)
  }

  if creds.EntityName == "" {
    logError("missing required field(s)")
    return nil, ErrDBMissingRequiredFields
  }

  // Query for Account via Account Email:
  var account struct{
    ID           uuid.UUID   `json:"id"`
    EntityID     uuid.UUID   `json:"entity_id"`
    PasswordHash string      `json:"password_hash"`
    Role         domain.Role `json:"role"`
  }

  if err := pg.db.QueryRow(
    ctx,
    `SELECT id, entity_id, password_hash, role accounts WHERE email = $1`,
    creds.Email,
  ).Scan(
    &account.ID,
    &account.EntityID,
    &account.PasswordHash,
    &account.Role,
  ); err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("failed to query user by email: " + err.Error())
      return nil, ErrDBFailedToQuery
    }

    logError("unknown error occurred: " + err.Error())
    return nil, err
  }

  // Validate Password and PasswordHash
  if valid := domain.ValidatePassword(
    creds.Password, 
    account.PasswordHash,
  ); !valid {
    logError("Account sign in failed -- Invalid Password")
    return nil, ErrDBInvalidPassword
  }

  // Generate JWT Tokens
  authTokens, err := domain.GenerateJWTTokens(
    account.ID,
    account.EntityID,
    account.Role,
  )
  if err != nil {
    logError("failed to create jwt tokens for successful account login")
    return nil, err
  }

  if err := pg.StoreRefreshToken(
    ctx, 
    account.ID,
    authTokens.RefreshToken,
  ); err != nil {}

  return authTokens, nil
}

// StoreRefreshToken - Tokes a newly created Refresh Token and Upserts it into the tokens DB table.
func(pg *PGRepo) StoreRefreshToken(
  ctx     context.Context, 
  acc_id  uuid.UUID,
  token   domain.Token,
) error {
  var logError = func(e string) {
    log.WithFields(log.Fields{
      "acc_id": acc_id,
    }).Error("StoreRefreshToken: " + e)
  }

  // Begin new DB Transaction:
  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("failed to begin DB transaction: " + err.Error())
    return ErrDBFailedToBeginTX
  }
  defer tx.Rollback(ctx)

  // Attempt to Upsert newly created Refresh Token:
  if _, err := tx.Exec(
    ctx,
    `INSERT INTO tokens (acc_id, refresh_token, expires_at)
     VALUES ($1, $2, $3)
     ON CONFLICT (acc_id, refresh_token) DO UPDATE
     SET expires_at = EXCLUDED.expires_at, updated_at = CURRENT_TIMESTAMP`,
     acc_id,
     token.SignedToken,
     token.Expiration,
  ); err != nil {
    logError("failed to upsert refresh token: " + err.Error())
    return ErrDBFailedToInsert
  }

  // Attempt to commit DB Transaction:
  if err := tx.Commit(ctx); err != nil {
    logError("Failed to commit DB Transaction: " + err.Error())
    return ErrDBFailedToCommitTX
  }

  return nil
}

// ValidateRefreshToken - Validates an Refresh Token by querying 'tokens' and detecting if expired.
func( pg *PGRepo) ValidateRefreshToken(
  ctx    context.Context,
  acc_id uuid.UUID,
  token  string,
) error {
  var logError = func(f string, data ...any) {
    log.WithFields(log.Fields{
      "acc_id": acc_id,
    }).Error(fmt.Sprintf(
      "ValidateRefreshToken:" + f,
      data...,
    ))
  }
  
  var dbToken string
  var expiresAt time.Time

  // Query for Account Refresh Token:
  if err := pg.db.QueryRow(
    ctx,
    `SELECT refresh_token, expires_at 
     FROM tokens 
     WHERE acc_id = $1 AND refresh_token = $2`,
    acc_id, 
    token,
  ).Scan(
    &dbToken, 
    &expiresAt,
  ); err != nil {
    logError("failed to query for refresh token: %v", err)
    return ErrDBFailedToQuery
  }

  // Check Refresh Token Expiration:
  if expiresAt.Before(time.Now()){
    logError("Refresh Token is Expired")
    return domain.ErrTokenExpired
  }

  return nil
}
