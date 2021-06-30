package handler

import (
	"context"
	"html/template"
	"net/http"

	hydraAdmin "github.com/ory/hydra-client-go/client/admin"
	"github.com/pkg/errors"
	// "github.com/imrushi/oauth2-openid-hydra/cmd/authgf/repouser"
)

type Handler struct {
	HydraAdmin hydraAdmin.ClientService
	// UserRepo   repouser.Repository
	Um UserManager
}

type UserManager interface {
	authenticator
	oidcClaimsFinder
}

type authenticator interface {
	Authenticate(ctx context.Context, username, password string) (ok bool, err error)
}

type oidcClaimsFinder interface {
	FindOIDCClaims(ctx context.Context, username string) (map[string]interface{}, error)
}

// renderTemplate is a convenience helper for rendering templates.
func renderTemplate(w http.ResponseWriter, id string, d interface{}) bool {
	if t, err := template.New(id).ParseFiles("../frontend/templates/" + id); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	} else if err := t.Execute(w, d); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	}
	return true
}
