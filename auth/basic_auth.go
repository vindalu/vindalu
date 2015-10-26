package auth

import (
	"fmt"
	"net/http"

	httpauth "github.com/abbot/go-http-auth"
)

/*
	Auth using htpasswd file
*/
type HttpBasicAuthenticator struct {
	//secretFunc httpauth.SecretProvider
	basicAuth *httpauth.BasicAuth
}

func NewHttpBasicAuthenticator(htpasswdFile string) *HttpBasicAuthenticator {
	return &HttpBasicAuthenticator{
		//secretFunc: httpauth.HtpasswdFileProvider(htpasswdFile),
		basicAuth: httpauth.NewBasicAuthenticator("vindaloo", httpauth.HtpasswdFileProvider(htpasswdFile)),
	}
}

func (hba *HttpBasicAuthenticator) AuthenticateRequest(r *http.Request) (username string, cacheHit bool, err error) {
	// Always true as it's local
	cacheHit = true

	username = hba.basicAuth.CheckAuth(r)
	if len(username) < 1 {
		err = fmt.Errorf("Auth failed!")
	}
	return
}
