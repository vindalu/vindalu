package auth

import (
	"time"
)

const DEFAULT_TTL int64 = 7200 // in seconds

type credentials struct {
	Username string `json:"username"`
	UserDN   string `json:"user_dn"`

	password string
	// Time cred struct was created
	CreatedTime int64 `json:"created_time"`
}

func NewCredentials(username, userDn, password string) *credentials {
	return &credentials{username, userDn, password, time.Now().Unix()}
}

type CredentialsCache struct {
	TTL   int64
	cache map[string]*credentials
}

func NewCredentialsCache(ttl int64) (cc *CredentialsCache) {
	cc = &CredentialsCache{cache: map[string]*credentials{}, TTL: ttl}

	if cc.TTL < 1 {
		cc.TTL = DEFAULT_TTL
	}

	return
}

func (cc *CredentialsCache) CheckCreds(username, password string) bool {
	v, ok := cc.cache[username]
	if !ok {
		return false
	}
	// Check expiration
	elapsed := time.Now().Unix() - v.CreatedTime
	if elapsed >= cc.TTL {
		cc.expireCacheCred(username)
		return false
	}
	if v.Username == username && v.password == password {
		return true
	}
	return false
}

func (cc *CredentialsCache) CacheCreds(creds *credentials) {
	_, ok := cc.cache[creds.Username]
	if ok {
		cc.expireCacheCred(creds.Username)
	}
	cc.cache[creds.Username] = creds
}

func (cc *CredentialsCache) expireCacheCred(username string) {
	delete(cc.cache, username)
}
