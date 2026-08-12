package main

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

type Config struct {
	Host                    string
	SPCert                  string
	SPKey                   string
	SamlIDPMetadataProvider string
}

func SetupSamlSP(c Config) *samlsp.Middleware {
	if c.SPCert == "" {
		log.Fatal("No SAML SP Cert provided")
	}
	if c.SPKey == "" {
		log.Fatal("No SAML SP Key provided")
	}
	if c.SamlIDPMetadataProvider == "" {
		log.Fatal("No SAML IDP Metadata url provided")

	}

	keyPair, err := tls.X509KeyPair([]byte(c.SPCert), []byte(c.SPKey))
	if err != nil {
		log.Fatalf("Failed to load SAML keypair: %s", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse leaf cert: %s", err)
	}

	idpMetadataURL, err := url.Parse(c.SamlIDPMetadataProvider)
	if err != nil {
		log.Fatalf("Failed to parse idpMetadataURL: %s", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		log.Fatalf("Failed to fetch idpMetadata: %s", err)
	}

	rootUrl, err := url.Parse(c.Host)
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

	samlSP.Session = &SessionProviderViaPlugin{}

	return samlSP
}

type SessionProviderViaPlugin struct {
}

// CreateSession is called when we have received a valid SAML assertion and
// should create a new session and modify the http response accordingly, e.g. by
// setting a cookie.
func (s *SessionProviderViaPlugin) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	panic("unimplemented")
}

// DeleteSession is called to modify the response such that it removed the current
// session, e.g. by deleting a cookie.
func (s *SessionProviderViaPlugin) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	panic("unimplemented")
}

// GetSession returns the current Session associated with the request, or
// ErrNoSession if there is no valid session.

func (s SessionProviderViaPlugin) GetSession(r *http.Request) (samlsp.Session, error) {
	panic("unimplemented")
}
