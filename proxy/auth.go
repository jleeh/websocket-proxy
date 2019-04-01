package proxy

import (
	"net/http"
)

const (
	AuthParam  = "param"
	AuthHeader = "header"
)

// Auth is the generic interface for how the client passes in their
// API key for authentication
type Auth interface {
	Authenticate(*http.Request, KeyManager) bool
}

// NewAuth returns a pointer of an Auth implementation based on the
// type that was passed in
func NewAuth(authType string) Auth {
	var a Auth
	switch authType {
	case AuthParam:
		a = &Param{}
		break
	case AuthHeader:
		a = &Header{}
	}
	return a
}

// Param is the Auth implementation that requires `apikey` in GET parameters
// to be set with the key
type Param struct{}

// Authenticate takes the `apikey` in the GET param and then authenticates it
// with the KeyManager
func (p *Param) Authenticate(r *http.Request, km KeyManager) bool {
	key := r.URL.Query().Get("apikey")
	return km.ValidateKey(key)
}

// Header is the Auth implementation that requires `X-API-KEY` to be set
// in the request headers
type Header struct{}

// Authenticate takes the `X-API-KEY` in the request headers and then authenticates
// it with the KeyManager
func (p *Header) Authenticate(r *http.Request, km KeyManager) bool {
	key := r.Header.Get("X-API-KEY")
	return km.ValidateKey(key)
}
