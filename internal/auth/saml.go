package auth

import (
	"make-backend/internal/database"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

type SCSSessionProvider struct {
	SessionManager *scs.SessionManager
	Store          *database.Store
}

func (p SCSSessionProvider) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	email := "placeholder" // TODO get email from assertion

	user, err := p.Store.Users.GetUserByEmail(r.Context(), email)
	if err != nil {
		return err
	}

	p.SessionManager.Put(r.Context(), "user_id", user.Id)

	return nil
}

func (p SCSSessionProvider) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	return p.SessionManager.Destroy(r.Context())
}

type dummysession struct{}

func (p SCSSessionProvider) GetSession(r *http.Request) (samlsp.Session, error) {
	if p.SessionManager.Exists(r.Context(), "user_id") {
		return dummysession{}, nil
	}

	return nil, samlsp.ErrNoSession
}
