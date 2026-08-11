package auth

import "net/rpc"

type AuthProvider interface {
	// Info() PluginInfoResponse
	GetLoginURL() string
}

type ToBase interface {
	NewUser(email string) (int, error)
	UserLoggedIn(userID int) error
	UserLoggedOut(userID int) error
}

type AuthPlugin struct {
	// Impl Injection
	Impl AuthProvider
}

// Here is an implementation that talks over RPC
type AuthProviderRPC struct{ client *rpc.Client }

// var _ AuthProvider = &AuthProviderRPC{}
