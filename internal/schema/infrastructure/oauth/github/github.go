package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/TylerAldrich814/Fidicus/internal/shared/config"
	"github.com/google/go-github/v55/github"
)

var (
  GithubSecret = config.GetEnv("GITHUB_WEBHOOK_SECRET", "")
)

// GithubWebhookHandler - Processess Github Webhook events
func GithubWebhookHandler(w http.ResponseWriter, r *http.Request) {
  payload := r.URL.Query().Get("payload")
  if payload == "" {
    http.Error(w, "payload is missing", http.StatusBadRequest)
    return
  }

  if GithubSecret == "" {
    http.Error(w, "internal error: missing github secret", http.StatusInternalServerError)
    return
  }

  if err := github.ValidateSignature("<TODO>", []byte(payload), []byte(GithubSecret)); err != nil {
    http.Error(w, "failed to validate github secret", http.StatusUnauthorized)
    return
  }
}

// ValidateSignature -- Extracts the Gihub Secret from URL Header.
// Using Gihub Secret, we take the signature, hash it via HMAC, and
// finally compare that is is equivalent to the Secret stored locally.
func ValidateSignature(r *http.Request) bool {
  signature := r.Header.Get("X-Hub-Signature-256")
  if signature == "" {
    return false
  }

  body, _ := io.ReadAll(r.Body)
  defer r.Body.Close()
  r.Body = io.NopCloser(strings.NewReader(string(body)))

  mac := hmac.New(sha256.New, []byte(GithubSecret))
  mac.Write(body)
  expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))


  return hmac.Equal([]byte(signature), []byte(expected))
}

func validateSchemaFromRepo(
  ctx context.Context,
  e *github.PushEvent,
) error {

  return nil
}
