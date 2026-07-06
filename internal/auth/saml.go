package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"make-backend/internal/database"
	"net/http"
	"net/url"
	"strconv"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func SetupSamlSP(store *database.Store) *samlsp.Middleware {
	keyPair, err := tls.LoadX509KeyPair("", "")
	if err != nil {
		log.Fatalf("Failed to load SAML keypair: %s", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse leaf cert: %s", err)
	}

	idpMetadataURL, err := url.Parse("")
	if err != nil {
		log.Fatalf("Failed to parse idpMetadataURL: %s", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		log.Fatalf("Failed to fetch idpMetadata: %s", err)
	}

	rootUrl, err := url.Parse("")
	if err != nil {
		log.Fatalf("Failed to parse root url: %s", err)
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:         *rootUrl,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	if err != nil {
		log.Fatalf("Failed to create samlSP: %s", err)
	}

	samlSP.Session = CustomSessionProvider{
		CookieSessionProvider: samlSP.Session.(samlsp.CookieSessionProvider),
		Store:                 store,
	}

	return samlSP
}

type CustomSessionProvider struct {
	CookieSessionProvider samlsp.CookieSessionProvider
	Store                 *database.Store
}

func (c CustomSessionProvider) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	username := "TODO" // assertion.ID

	user, err := c.Store.Users.GetUserByUsername(r.Context(), username)
	if err != nil {
		return err
	}

	user_id := strconv.Itoa(user.Id)

	customAttribute := saml.Attribute{
		Name:   "make_user_id",
		Values: []saml.AttributeValue{{Value: user_id}},
	}

	if len(assertion.AttributeStatements) > 0 {
		assertion.AttributeStatements[0].Attributes = append(assertion.AttributeStatements[0].Attributes, customAttribute)
	} else {
		assertion.AttributeStatements = []saml.AttributeStatement{
			{
				Attributes: []saml.Attribute{customAttribute},
			},
		}
	}

	return c.CookieSessionProvider.CreateSession(w, r, assertion)
}

func (c CustomSessionProvider) GetSession(r *http.Request) (samlsp.Session, error) {
	return c.CookieSessionProvider.GetSession(r)
}

func (c CustomSessionProvider) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	return c.CookieSessionProvider.DeleteSession(w, r)
}
