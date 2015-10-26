package auth

import (
	//"io/ioutil"
	"testing"
)

var (
	testTokenUser = "admin"
	testSignKey   = []byte(`testkey`)
	testTokenStr  string
)

func Test_GetNewToken_default_ttl(t *testing.T) {
	var (
		token = GetNewToken(testTokenUser, 0)
		err   error
	)
	t.Logf("New token: %#v\n", token)

	testTokenStr, err = token.SignedString(testSignKey)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("Token string: %s", testTokenStr)
}

func Test_GetNewToken_1h_ttl(t *testing.T) {
	var (
		token = GetNewToken(testTokenUser, 1)
		err   error
	)
	t.Logf("New token: %#v\n", token)

	testTokenStr, err = token.SignedString(testSignKey)
	if err != nil {
		t.Fatalf("%s", err)
	}

	t.Logf("Token string: %s", testTokenStr)
}

func Test_GetTokenFromString(t *testing.T) {
	token, err := GetTokenFromString(testTokenStr, testSignKey)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if !token.Valid {
		t.Fatalf("Token should be valid")
	}

	t.Logf("Parsed token: %#v", token)
}

func Test_GetTokenFromString_Error(t *testing.T) {
	if _, err := GetTokenFromString(testTokenStr, testSignKey[1:]); err == nil {
		t.Fatalf("Should have failed parsing token")
	}
}
