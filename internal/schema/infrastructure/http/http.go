package http

import (
	"net/http"

	"github.com/TylerAldrich814/Fidicus/internal/shared/role"
	"github.com/TylerAldrich814/Fidicus/internal/shared/middleware"
	"github.com/TylerAldrich814/Fidicus/internal/schema/application"
	"github.com/gorilla/mux"
)

// SchemaHTTPHandler defines a structure for handling all HTTP requests
// relating to Schema Management, Storage, and Validation.
type SchemaHTTPHandler struct {
  service *application.Service
}

// NewHTTPHandler - Creates a new SchemaHTTPHandler instance.
func NewHTTPHandler(
  service *application.Service,
) *SchemaHTTPHandler {
  return &SchemaHTTPHandler{ service }
}

func(s *SchemaHTTPHandler) RegisterRoutes(r mux.Router) error {
  schema := r.PathPrefix("/schemas").Subrouter()
  schema.Use(middleware.AuthMiddleware)

  schema.Handle(
    "/upload",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(s.UploadSchemas),
      role.AccessRoleAccount,
    ),
  ).Methods("PUT")

  schema.Handle(
    "/delete",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(s.DeleteSchema),
      role.AccessRoleAccount,
    ),
  ).Methods("DELETE")

  schema.HandleFunc(
    "/download/{source}/{id}",
    s.GetSchemas,
  ).Methods("GET")

  schema.Handle(
    "/sync",
    middleware.RoleAuthMiddleware(
      http.HandlerFunc(s.SyncSchemas),
      role.AccessRoleAccount,
    ),
  ).Methods("POST")

  schema.HandleFunc(
    "/list",
    s.ListSchemas,
  ).Methods("GET")

  schema.HandleFunc(
    "/validate/{id}",
    s.Validate,
  ).Methods("GET")

  return nil
}

func(s *SchemaHTTPHandler) UploadSchemas(w http.ResponseWriter, r *http.Request) {


  if err := s.service.UploadSchema(r.Context()); err != nil {
    http.Error(w, "failed to upload schema", http.StatusBadRequest)
  }
}

func(s *SchemaHTTPHandler) DeleteSchema(w http.ResponseWriter, r *http.Request) {

}

func(s *SchemaHTTPHandler) GetSchemas(w http.ResponseWriter, r *http.Request){

}

func(s *SchemaHTTPHandler) SyncSchemas(w http.ResponseWriter, r *http.Request) {
}

func(s *SchemaHTTPHandler) ListSchemas(w http.ResponseWriter, r *http.Request) {

}

func(s *SchemaHTTPHandler) Validate(w http.ResponseWriter, r *http.Request) {

}
