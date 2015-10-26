package auth

import (
	"testing"
)

var (
	testUser  = "testuser"
	testPass  = "testpass"
	testCreds = NewCredentials(testUser, "", testPass)
	testCache = NewCredentialsCache(0)
)

func Test_NewCredentialsCache_TTL(t *testing.T) {
	ccache := NewCredentialsCache(400)
	if ccache.TTL != 400 {
		t.Fatalf("Failed to set ttl: %d", ccache.TTL)
	}
}

func Test_NewCredentialsCache(t *testing.T) {
	ccache := NewCredentialsCache(0)
	if ccache.TTL != DEFAULT_TTL {
		t.Fatalf("Failed to set default ttl: %d", ccache.TTL)
	}
}

func Test_CredentialsCache_CheckCreds(t *testing.T) {
	testCache.CacheCreds(testCreds)
	if !testCache.CheckCreds(testUser, testPass) {
		t.Fatalf("Failed to check creds")
	}
}
