package application

import (
	"context"

	"github.com/TylerAldrich814/Fidicus/internal/schema/domain"
)

type Service struct {
  blob domain.SchemaBlobRepository
  grph domain.SchemaGraphRepository
  psql domain.SchemaSQLRepository
}

func NewService(
  blob domain.SchemaBlobRepository,
  grph domain.SchemaGraphRepository,
  psql domain.SchemaSQLRepository,
) *Service {
  return &Service{ blob, grph, psql, }
}

func(s *Service) Shutdown() error {
  if s.blob != nil {
    if err := s.blob.Shutdown(); err != nil {
      return err
    }
  }
  if s.grph != nil {
    if err := s.grph.Shutdown(); err != nil {
      return err
    }
  }
  if s.psql != nil {
    if err := s.psql.Shutdown(); err != nil {
      return err
    }
  }
  return nil
}

func(s *Service) UploadSchema(
  ctx context.Context,
) error {

  return nil
}
