package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"

	"github.com/TylerAldrich814/Fidicus/internal/shared/utils"
	"github.com/TylerAldrich814/Fidicus/internal/shared/role"
	"github.com/TylerAldrich814/Fidicus/internal/shared/users"
	"github.com/TylerAldrich814/Fidicus/internal/shared/jwt"
)

// PGRepo -- A Postgres wrapper that implements AuthRepository
type PGRepo struct {
  db *pgxpool.Pool
}

// New -- Creates a new PGRepo instance.
func NewAuthRepo(
  ctx context.Context,
  dsn string,
)( *PGRepo, error ){
  var pushLog = utils.NewLogHandlerFunc(
    "NewAuthRepo",
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
    return nil, ErrPGSQLConfigParse
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
    return nil, ErrDBFailedCreation
  }

  if err := pool.Ping(ctx); err != nil {
    pushLog(
      utils.LogErro,
      "failed to ping database with newly created postgres pool: %s",
      err.Error(),
    )
    return nil, ErrDBFailedPing
  }
  
  return &PGRepo{ pool }, nil
}

// CreateEntity -- Creates a new Fidicus Entity with Root Privileges.
//
// Potential Errors:
//   - ErrDBFailedToQuery
//   - ErrDBEntityAlreadyExists
//   - ErrDBAccountAlreadyExists
//   - ErrDBInternalFailure
//   - ErrDBFailedToBeginTX
//   - ErrDBEntityAlreadyExists
//   - ErrDBFailedToInsert
//   - ErrDBFailedToCommitTX
func(pb *PGRepo) CreateEntity(
  ctx        context.Context, 
  entityReq  users.EntitySignupReq,
  accountReq users.AccountSignupReq,
)( users.EntityID, users.AccountID, error) {
  var logError = func(f string, data ...any) {
    log.WithFields(log.Fields{
      "entity_name"   : entityReq.Name,
      "account_email" : accountReq.Email,
    }).Error(fmt.Sprintf(
      "CreateEntity: "+f,
      data...,
    ))
  }

  // ->> Verify that both Entity and Account don't exist yet.
  var exists bool
  qErr := pb.db.QueryRow(
    ctx,
    `SELECT EXISTS(SELECT 1 FROM entities WHERE name = $1)`,
    entityReq.Name,
  ).Scan(&exists)

  if qErr != nil {
    logError("Failed to query entities: %v", qErr)
    return users.NilEntity(), users.NilAccount(), ErrDBFailedToQuery
  }
  if exists {
    logError("entity already exists")
    return users.NilEntity(), users.NilAccount(), ErrDBEntityAlreadyExists
  }

  qErr = pb.db.QueryRow(
    ctx,
    `SELECT EXISTS(SELECT 1 FROM accounts WHERE email = $1)`,
    accountReq.Email,
  ).Scan(&exists)
  if qErr != nil {
    logError("Failed to query accounts: %v", qErr)
    return users.NilEntity(), users.NilAccount(), ErrDBFailedToQuery
  }
  if exists {
    logError("account already exists")
    return users.NilEntity(), users.NilAccount(), ErrDBAccountAlreadyExists
  }

  // ->> Generate IDs
  entityID  := users.NewEntityID()
  accountID := users.NewAccountID()

  entity := users.Entity {
    ID          : entityID,
    Name        : entityReq.Name,
    Description : entityReq.Description,
    AccountIDs  : []users.AccountID{accountID},
    CreatedAt   : time.Now(),
    UpdatedAt   : time.Now(),
  }

  hashPassw, err := users.HashPassword(accountReq.Passw)
  if err != nil {
    logError("failed to hash password: %v", err)
    return users.NilEntity(), users.NilAccount(), ErrDBInternalFailure
  }
  account := users.Account {
    ID              : accountID,
    EntityID        : entityID,
    Email           : accountReq.Email,
    PasswHash       : hashPassw,
    Role            : role.AccessRoleEntity,
    FirstName       : accountReq.FirstName,
    LastName        : accountReq.LastName,
    CellphoneNumber : accountReq.CellphoneNumber,
  }

  tx, err := pb.db.Begin(ctx)
  if err != nil {
    logError("failed to create transaction: %v", err)
    return users.NilEntity(), users.NilAccount(), ErrDBFailedToBeginTX
  }
  defer tx.Rollback(ctx)

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
      return users.NilEntity(), users.NilAccount(), ErrDBEntityAlreadyExists
    }
    logError("entity creation failed: " + err.Error())
    return  users.NilEntity(), users.NilAccount(), ErrDBFailedToInsert
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
      return users.NilEntity(), users.NilAccount(), ErrDBAccountAlreadyExists
    }
    logError("account creation failed: " + err.Error())
    return users.NilEntity(), users.NilAccount(), ErrDBFailedToInsert
  }

  if err := tx.Commit(ctx); err != nil {
    logError("Failed to commit DB transaction: " + err.Error())
    return users.NilEntity(), users.NilAccount(), ErrDBFailedToCommitTX
  }

  return users.EntityID(entityID), users.AccountID(accountID), nil
}

// CreateAccount - Creates a new Entity Account.
//
// Potential Errors:
//   - ErrDBUnauthorized
//   - ErrDBFailedToQuery
//   - ErrDBEntityNotFound
//   - ErrDBFailedToBeginTX
//   - ErrDBInternalFailure
//   - ErrDBAccountAlreadyExists
//   - ErrDBFailedToInsert
//   - ErrDBFailedToCommitTX
//   - ErrDBMissingRequiredFields
func(pg *PGRepo) CreateAccount(
  ctx        context.Context, 
  accountReq users.AccountSignupReq,
)( users.AccountID, error){
  var logError = func(f string, data ...any) {
    log.WithFields(log.Fields{
      "eid"   : accountReq.EntityID,
      "ename" : accountReq.EntityName,
      "email" : accountReq.Email,
      "fName" : accountReq.FirstName,
      "lName" : accountReq.LastName,
      "role"  : accountReq.Role,
    }).Error(fmt.Sprintf("CreateAccount: "+f, data...))
  }

  // <TODO> :: Should we allow access_role_entity accounts create fellow role_account_entitys ??
  if accountReq.Role == role.AccessRoleEntity {
    logError("Access Role Entity can only be created during Entity creation")
    return users.NilAccount(), ErrDBUnauthorized
  }

  var err error
  eid := accountReq.EntityID
  if eid == users.NilEntity() {
    if accountReq.EntityName == "" {
      logError(
        "missing entity information: ID or Name are required",
      )
      return users.NilAccount(), ErrDBMissingRequiredFields
    }
    eid, err = pg.GetEntityIDByName(ctx, accountReq.EntityName)
    if err != nil {
      logError(
        "entity name does not correlate with any entity in system: %s - %s", 
        accountReq.EntityName, 
        err.Error(),
      )
      return users.NilAccount(), ErrDBEntityNotFound
    }
  } 

  // ->> Check if Entity Exists:
  var exists bool
  if qErr := pg.db.QueryRow(
    ctx,
    `SELECT EXISTS(SELECT 1 FROM entities WHERE id = $1)`,
    eid,
  ).Scan(&exists); qErr != nil {
    logError("failed to query entities: %v", qErr)
    return users.NilAccount(), ErrDBFailedToQuery
  }

  if !exists {
    logError("entity doesn't exist")
    return users.NilAccount(), ErrDBEntityNotFound
  }

  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("failed to create new transaction: %v", err)
    return users.NilAccount(), ErrDBFailedToBeginTX
  }

  // ->> Create New Account Data
  passwHash, err := users.HashPassword(accountReq.Passw)
  if err != nil {
    logError("failed to hash password: %v", err)
    return users.NilAccount(), ErrDBInternalFailure
  }

  accountID := uuid.New()
  account := users.Account {
    ID              : users.NewAccountID(),
    EntityID        : accountReq.EntityID,
    Email           : accountReq.Email,
    PasswHash       : passwHash,
    Role            : accountReq.Role,
    FirstName       : accountReq.FirstName,
    LastName        : accountReq.LastName,
    CellphoneNumber : accountReq.CellphoneNumber,
    CreatesAt       : time.Now(),
    UpdatedAt       : time.Now(),
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
    eid,
    account.Email, 
    account.PasswHash,
    account.Role,
    account.FirstName,
    account.LastName,
    account.CellphoneNumber,
    account.CreatesAt,
    account.UpdatedAt,
  )
  if err != nil {
    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) && pgErr.Code == "23505" {
      logError("account already exists")
      return users.NilAccount(), ErrDBAccountAlreadyExists
    }
    logError("failed to insert new Account: %v", err)
    return  users.NilAccount(), ErrDBFailedToInsert
  }

  // Commit Changes:
  if err := tx.Commit(ctx); err != nil {
    logError("failed to commit DB Transaction: %v", err)
    return users.NilAccount(), ErrDBFailedToCommitTX
  }

  return users.AccountID(accountID), nil
}

// RemoveEntityByID - <TEMP> Removes an entity via it's ID
//
// Potential Errors:
//   - ErrDBFailedToBeginTX
//   - ErrDBFailedToDeleteEntity
//   - ErrDBFailedToCommitTX
func(pg *PGRepo) RemoveEntityByID(
  ctx context.Context, 
  id users.EntityID,
) error {
  var logError = func(f string, args ...any) {
    log.WithFields(log.Fields{
      "id": id,
    }).Error(fmt.Sprintf("RemoveEntityByID: "+f, args...))
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
//
// Potential Errors:
//   - ErrDBFailedToBeginTX
//   - ErrDBFailedToDeleteEntity
//   - ErrDBFailedToCommitTX
func(pg *PGRepo) RemoveEntityByName(
  ctx        context.Context,
  entityName string,
) error {
  eid, err := pg.GetEntityIDByName(ctx, entityName)
  if err != nil {
    log.Error("RemoveEntityByName: Failed to query Entity ID: " + err.Error())
    return err
  }
  return pg.RemoveEntityByID(
    ctx, 
    eid,
  )
}

// RemoveAccountByID    - <TEMP> Removes an Account via it's ID
// 
// Potential Errors:
//   - ErrDBFailedToBeginTX
//   - ErrDBFailedToDeleteAccount
//   - ErrDBFailedToCommitTX
func(pg *PGRepo) RemoveAccountByID(
  ctx context.Context, 
  id  users.AccountID,
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
//
// Potential Errors:
//   - ErrDBFailedToBeginTX
//   - ErrDBFailedToDeleteAccount
//   - ErrDBFailedToCommitTX
func(pg *PGRepo) RemoveAccountByEmail(
  ctx context.Context, 
  email string,
) error {
  id, err := pg.GetAccountIDByEmail(ctx, email)
  if err != nil {
    return err
  }

  return pg.RemoveAccountByID(ctx, id)
}

// GetEntityIDByName - Query and returns an Entitys ID via it's name.
//
// Potential Errors:
//   - ErrDBEntityNotFound
//   - ErrDBInternalFailure
func(pg *PGRepo) GetEntityIDByName(
  ctx  context.Context, 
  name string,
)( users.EntityID, error ){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "name": name,
    }).Error("GetEntityIDByName: " + e)
  }

  var entityID users.EntityID

  err := pg.db.QueryRow(
    ctx,
    `SELECT id FROM entities WHERE name = $1`,
    name,
  ).Scan(&entityID)
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("No Rows with name found")
      return users.NilEntity(), ErrDBEntityNotFound
    }
    logError("Unknown error occurred: " + err.Error())
    return users.NilEntity(), ErrDBInternalFailure
  }

  return entityID, nil
}

// GetAccountIdByName - Query and returns an Accounts ID via it's Email.
// Potential Errors:
//   - ErrDBAccountNotFound
//   - ErrDBInternalFailure
func(pg *PGRepo) GetAccountIDByEmail(
  ctx   context.Context, 
  email string,
)( users.AccountID, error ){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "email": email,
    }).Error("GetAccountIdByName: " + e)
  }

  var accountID users.AccountID

  if err := pg.db.QueryRow(
    ctx,
    `SELECT id FROM accounts WHERE email = $1`,
    email,
  ).Scan(&accountID); err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("no Rows with email found")
      return users.NilAccount(), ErrDBAccountNotFound
    }
    logError("Unknown error occurred: " + err.Error())
    return users.NilAccount(), ErrDBInternalFailure
  }

  return accountID, nil
}

// AccountSignin - Before signing account in. First detects if account exists 
// then checks to make sure account is a Subaccount of Entity.
// 
// Potential Errors:
//   - ErrDBMissingRequiredFields
//   - ErrDBFailedToQuery
//   - ErrDBInternalFailure
//   - ErrDBInvalidPassword
//   - ErrDBFailedToInsert
func(pg *PGRepo) AccountSignin(
  ctx       context.Context, 
  signinReq users.AccountSigninReq,
)( jwt.Token, jwt.Token, error){
  var logError = func(e string){
    log.WithFields(log.Fields{
      "Entity"    : signinReq.EntityName,
      "email"     : signinReq.Email,
      "pw_exists" : len(signinReq.Passw) != 0,
    }).Error("AccountSignin: " + e)
  }

  if signinReq.EntityName == "" {
    logError("missing required field(s)")
    return jwt.Token{}, jwt.Token{}, ErrDBMissingRequiredFields
  }

  // ->> Query for Account via Account Email:
  var account struct{
    EntityID     users.EntityID  `json:"entity_id"`
    ID           users.AccountID `json:"id"`
    PasswordHash string           `json:"password_hash"`
    Role         role.Role      `json:"role"`
  }

  if err := pg.db.QueryRow(
    ctx,
    `SELECT id, entity_id, password_hash, role FROM accounts WHERE email = $1`,
    signinReq.Email,
  ).Scan(
    &account.ID,
    &account.EntityID,
    &account.PasswordHash,
    &account.Role,
  ); err != nil {
    if errors.Is(err, pgx.ErrNoRows){
      logError("failed to query user by email: " + err.Error())
      return jwt.Token{}, jwt.Token{}, ErrDBFailedToQuery
    }

    logError("unknown error occurred: " + err.Error())
    return jwt.Token{}, jwt.Token{}, ErrDBInternalFailure
  }

  // ->> Validate Password and PasswordHash
  if valid := users.ValidatePassword(
    signinReq.Passw, 
    account.PasswordHash,
  ); !valid {
    logError("Account sign in failed -- Invalid Password")
    return jwt.Token{}, jwt.Token{}, ErrDBInvalidPassword
  }

  // ->> Generate JWT Tokens
  newAccessToken, err := jwt.GenerateAccessToken(
    account.ID,
    account.EntityID,
    account.Role,
  )
  if err != nil || newAccessToken.SignedToken == "" {
    return jwt.Token{}, jwt.Token{}, jwt.ErrTokenGenFailed
  }

  newRefreshToken, err := jwt.GenerateRefreshToken(
    account.ID,
    account.EntityID,
    account.Role,
  )
  if err != nil || newRefreshToken.SignedToken == "" {
    return jwt.Token{}, jwt.Token{}, jwt.ErrTokenGenFailed
  }

  // ->> Store newly created RefreshToken.
  if err := pg.StoreRefreshToken(
    ctx, 
    account.ID,
    newRefreshToken,
  ); err != nil {
    logError("AccountSignin: " + err.Error())
    return jwt.Token{}, jwt.Token{}, err
  }

  return newAccessToken, newRefreshToken, nil
}

// AccountSignout - Signs the account out of Fidicus. Removing their refresh token from  
// our tokens DB. By the time this function is called. The user should have already been
// validated via their provided access token.
func(pg *PGRepo) AccountSignout(
  ctx       context.Context,
  accountID users.AccountID,
) error {
  var logError = func(f string, args ...any) {
    log.WithFields(log.Fields{
      "account_id": accountID,
    }).Error(fmt.Sprintf("AccountSignout: "+f, args...))
  }

  tx, err := pg.db.Begin(ctx)
  if err != nil {
    logError("failed to begin DB transaction: %s", err.Error())
    return ErrDBFailedToBeginTX
  }

  if _, err := tx.Exec(
    ctx,
    `DELETE FROM tokens WHERE account_id = $1`,
    accountID,
  ); err != nil {
    logError("failed to remove user access token from tokens table: %s", err.Error())
    return ErrDBFailedToDeleteToken
  }

  if err := tx.Commit(ctx); err != nil {
    logError("failed to commit DB transaction: %s", err.Error())
    return ErrDBFailedToCommitTX
  }

  return nil
}

// StoreRefreshToken - Tokes a newly created Refresh Token and Upserts it into the tokens DB table.
//
// Potential Errors:
//   - ErrDBFailedToBeginTX
//   - ErrDBFailedToInsert
//   - ErrDBFailedToCommitTX
func(pg *PGRepo) StoreRefreshToken(
  ctx       context.Context, 
  accountID users.AccountID,
  token     jwt.Token,
) error {
  var logError = func(e string) {
    log.WithFields(log.Fields{
      "accountID": accountID,
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
    `INSERT INTO tokens (account_id, refresh_token, expires_at, updated_at)
    VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
    ON CONFLICT (refresh_token) DO UPDATE
    SET expires_at = EXCLUDED.expires_at, updated_at = CURRENT_TIMESTAMP`,
    accountID,
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
//
// Potential Errors:
//   - ErrDBFailedToQuery
//   - users.ErrTokenExpired
func( pg *PGRepo) ValidateRefreshToken(
  ctx       context.Context,
  accountID users.AccountID,
  token     string,
) error {
  var logError = func(f string, data ...any) {
    log.WithFields(log.Fields{
      "accountID": accountID,
    }).Error(fmt.Sprintf(
      "ValidateRefreshToken:" + f,
      data...,
    ))
  }
  var expiresAt time.Time

  // Query for Account Refresh Token:
  if err := pg.db.QueryRow(
    ctx,
    `SELECT expires_at 
     FROM tokens 
     WHERE account_id = $1 AND refresh_token = $2`,
    accountID, 
    token,
  ).Scan(
    &expiresAt,
  ); err != nil {
    logError("failed to query for refresh token: %v", err)
    return ErrDBFailedToQuery
  }

  // Check Refresh Token Expiration:
  if expiresAt.Before(time.Now()){
    logError("Refresh Token is Expired")
    return jwt.ErrTokenExpired
  }

  return nil
}

// RefreshToken - For creating a new Access Token, requires an accountID to verify account validity.
func(pg *PGRepo) CreateRefreshToken(
  ctx       context.Context, 
  entityID  users.EntityID,
  accountID users.AccountID,
  role      role.Role,
)( jwt.Token, jwt.Token, error ){

  var throwError = func(err error)(jwt.Token, jwt.Token, error){
    return jwt.Token{}, jwt.Token{}, err
  }

  // ->> Generate new JWT Tokens.
  newAccessToken, err := jwt.GenerateAccessToken(accountID, entityID, role)
  if err != nil {
    return throwError(err)
  }

  newRefreshToken, err := jwt.GenerateRefreshToken(accountID, entityID, role)
  if err != nil {
    return throwError(err)
  }

  // ->> Store new Refresh Token
  if err := pg.StoreRefreshToken(ctx, accountID, newRefreshToken); err != nil {
    return throwError(err)
  }

  return newAccessToken, newRefreshToken, nil
}

func(pg *PGRepo) Shutdown() error{
  if pg.db == nil {
    log.Warn("tried to shutdown postgres but postgres is already down.")
    return errors.New("Postgres wasn't running")
  } 

  log.Info("Shutting Auth Repo Down...")
  pg.db.Close()

  return nil
}
