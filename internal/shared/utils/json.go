package utils

import (
  "encoding/json"
  "net/http"
)

// WriteJson - Writes a data object to an HTTP Responses JSON Body.
func WriteJson(
  w      http.ResponseWriter,
  status int,
  data   any,
) {
  w.Header().
    Set("Content-Type", "application/json")

  w.WriteHeader(status)
  json.NewEncoder(w).Encode(data)
}

// ReadJson - Taking in an HTTP Request -- Attempts to extract the JSON object from the request.
func ReadJson(
  r *http.Request,
  data interface{},
) error {
  return json.NewDecoder(r.Body).Decode(data)
}

// WriteError - Writes an JSON Error Message 
func WriteError(
  w      http.ResponseWriter,
  status int,
  err    string,
) {
  WriteJson(
    w,
    status,
    map[string]string{
      "error": err,
    },
  )
}
