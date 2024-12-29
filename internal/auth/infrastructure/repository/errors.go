package repository

import "errors"

var (
  ErrDBConfigParse           = errors.New("failed to parse database config")
  ErrDBFailedCreation        = errors.New("failed to create postgres database")
  ErrDBFailedPing            = errors.New("failed to ping postgres database")

  ErrDBFailedToBeginTX       = errors.New("failed to create new DB Transaction")
  ErrDBFailedToCommitTX      = errors.New("database transaction commit failed")

  ErrDBEntityAlreadyExists   = errors.New("attempted to create an entity that already exists")
  ErrDBAccountAlreadyExists  = errors.New("attempted to create an account that already exists")

  ErrDBEntityNotFound        = errors.New("queried entity doesn't exists")
  ErrDBAccountNotFound       = errors.New("queried account doesn't exists")

  ErrDBFailedToInsert        = errors.New("failed to insert into DB table")
  ErrDBFailedToQuery         = errors.New("failed to query for an account")
  ErrDBInvalidPassword       = errors.New("invalid password")

  ErrDBFailedToDeleteEntity  = errors.New("failed to delete entity")
  ErrDBFailedToDeleteAccount = errors.New("failed to delete entity accounts")

  ErrDBMissingRequiredFields = errors.New("DB Request missing required fields")
)
