package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"make-backend/internal/plugins/auth"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

type Config struct {
	BaseURL                 string
	SPCert                  string
	SPKey                   string
	SamlIDPMetadataProvider string
}

func (s *SAMLAuth) SetupSamlSP(c Config, sp *SessionProviderViaPlugin) (*samlsp.Middleware, error) {
	keyPair, err := tls.X509KeyPair([]byte(c.SPCert), []byte(c.SPKey))
	if err != nil {
		return nil, fmt.Errorf("failed to load SAML keypair: %w", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("Failed to parse leaf cert: %w", err)
	}

	idpMetadataURL, err := url.Parse(c.SamlIDPMetadataProvider)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse idpMetadataURL: %w", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch idpMetadata: %w", err)
	}

	rootUrl, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root url: %s", err)
	}

	samlSP, err := samlsp.New(samlsp.Options{
		URL:                *rootUrl,
		Key:                keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate:        keyPair.Leaf,
		IDPMetadata:        idpMetadata,
		DefaultRedirectURI: c.BaseURL + "/help",
	})
	metaU, _ := url.Parse(fmt.Sprintf("%s/metadata", c.BaseURL))
	acsU, _ := url.Parse(fmt.Sprintf("%s/acs", c.BaseURL))

	samlSP.ServiceProvider.MetadataURL = *metaU
	samlSP.ServiceProvider.AcsURL = *acsU

	if err != nil {
		return nil, fmt.Errorf("failed to create samlSP: %w", err)
	}
	samlSP.ServiceProvider.AuthnNameIDFormat = saml.EmailAddressNameIDFormat

	samlSP.Session = sp

	return samlSP, nil
}

type SessionProviderViaPlugin struct {
	cb auth.AuthCallbackProvider
}

func extractStringFromAttributeValues(values []saml.AttributeValue) string {
	for _, value := range values {
		return value.Value
	}
	return ""
}

func mapAssertionToCallback(assertion *saml.Assertion, cb *auth.UserLoginCallback) {
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			if attr.Name == "email" {
				cb.Email = extractStringFromAttributeValues(attr.Values)
			}
		}
	}
}

// CreateSession is called when we have received a valid SAML assertion and
// should create a new session and modify the http response accordingly, e.g. by
// setting a cookie.
func (s *SessionProviderViaPlugin) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	cb := auth.UserLoginCallback{}
	mapAssertionToCallback(assertion, &cb)
	resp, err := s.cb.UserLoggedIn(&cb)
	if err != nil {
		return err
	}
	for _, c := range resp.SetHeaders {
		w.Header().Add(c.Key, c.Value)
	}
	w.Header().Add("Location", resp.GotoUrl)

	return nil
}

// DeleteSession is called to modify the response such that it removed the current
// session, e.g. by deleting a cookie.
func (s *SessionProviderViaPlugin) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	log.Println("delete session")
	return nil

}

// GetSession returns the current Session associated with the request, or
// ErrNoSession if there is no valid session.

func (s SessionProviderViaPlugin) GetSession(r *http.Request) (samlsp.Session, error) {
	log.Println("get session")
	return nil, samlsp.ErrNoSession

}
