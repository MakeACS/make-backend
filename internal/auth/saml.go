package auth

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"make-backend/internal/database"
	"net/http"
	"net/url"
	"os"

	"github.com/alexedwards/scs/v2"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

func SetupSamlSP(store *database.Store, sessionManager *scs.SessionManager) *samlsp.Middleware {
	spCert := os.Getenv("SAML_SP_CERT")
	spKey := os.Getenv("SAML_SP_KEY")
	if spCert == "" {
		log.Fatal("No SAML SP Cert provided")
	}
	if spKey == "" {
		log.Fatal("No SAML SP Key provided")
	}

	keyPair, err := tls.X509KeyPair([]byte(spCert), []byte(spKey))
	if err != nil {
		log.Fatalf("Failed to load SAML keypair: %s", err)
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse leaf cert: %s", err)
	}

	idpMetadataURL, err := url.Parse(os.Getenv("SAML_SP_METADATA_URL"))
	if err != nil {
		log.Fatalf("Failed to parse idpMetadataURL: %s", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURL)
	if err != nil {
		log.Fatalf("Failed to fetch idpMetadata: %s", err)
	}

	rootHost := os.Getenv("HOST_URL")
	if rootHost == "" {
		log.Fatalf("No HOST_URL provided")
	}

	rootUrl, err := url.Parse(fmt.Sprintf("%s:%s", rootHost, os.Getenv("PORT")))
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

	samlSP.Session = SCSSessionProvider{
		SessionManager: sessionManager,
		Store:          store,
	}

	return samlSP
}

type SCSSessionProvider struct {
	SessionManager *scs.SessionManager
	Store          *database.Store
}

func (p SCSSessionProvider) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	username := "placeholder" // TODO get username from assertion

	user, err := p.Store.Users.GetUserByUsername(r.Context(), username)
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
