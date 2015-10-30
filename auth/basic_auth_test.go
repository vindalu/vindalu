package auth

import (
	"net/http"
	"testing"
)

func Test_NewNewHttpBasicAuthenticator(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v3/type/id", nil)
	req.SetBasicAuth("admin", "vindaloo")

	bauth := NewHttpBasicAuthenticator("../etc/htpasswd")
	if _, _, err := bauth.AuthenticateRequest(req); err != nil {
		t.Fatalf("%s", err)
	}
}

func Test_NewNewHttpBasicAuthenticator_error(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v3/type/id", nil)
	req.SetBasicAuth("admin", "vind")

	bauth := NewHttpBasicAuthenticator("../etc/htpasswd")
	if _, _, err := bauth.AuthenticateRequest(req); err == nil {
		t.Fatalf("Should have failed auth")
	}
}
