package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"log/slog"
	"make-backend/internal/database"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func SetupSamlSP(store *database.Store) *samlsp.Middleware {
	keyPair, err := tls.LoadX509KeyPair("certs/SPcert.pem", "certs/SPkey.pem")
	if err != nil {
		log.Fatalf("Failed to load SAML keypair: %s", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse leaf cert: %s", err)
	}

	idpMetadataURL, err := url.Parse("http://mocksaml.com/api/saml/metadata")
	if err != nil {
		log.Fatalf("Failed to parse idpMetadataURL: %s", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		log.Fatalf("Failed to fetch idpMetadata: %s", err)
	}

	rootUrl, err := url.Parse("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to parse root url: %s", err)
	}

	samlSP, err := samlsp.New(samlsp.Options{

		URL:         *rootUrl,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	samlSP.ServiceProvider.AuthnNameIDFormat = saml.EmailAddressNameIDFormat

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

var usernamePath = "urn:oid:0.9.2342.19200300.100.1.1"
var firstnamePath = "urn:oid:2.5.4.42"
var lastnamePath = "urn:oid:2.5.4.4"
var rolesPath = "urn:oid:1.3.6.1.4.1.4447.1.41"

func extractStringFromAttributeValues(values []saml.AttributeValue) string {
	for _, value := range values {
		return value.Value
	}
	return ""
}

func mapSamlTestToRit(assertion *saml.Assertion) string {
	var username string = ""
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if attr.Name == "email" {
				username = strings.Split(extractStringFromAttributeValues(attr.Values), "@")[0]
			}
		}
	}
	return username
}

func (c CustomSessionProvider) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	var username = mapSamlTestToRit(assertion)
	slog.Debug("Session creation", "username", username)

	user, err := c.Store.Users.GetUserByUsername(r.Context(), username)
	if err != nil {
		// slog.Error("")
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
