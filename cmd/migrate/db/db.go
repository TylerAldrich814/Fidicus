package db

type DB interface {
  Migrate(arg string) error
}
