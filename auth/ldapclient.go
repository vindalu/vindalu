package auth

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap"
)

type LDAPClient struct {
	LdapConn *ldap.Conn
	// Default search base
	DefaultSearchBase string
	// ldap uri
	URI          string
	UserBindDN   string
	UserBindPass string

	hostPort string

	cache *CredentialsCache
}

func NewLDAPClient(ldapUri, bindDn, bindPass string, cacheTTL int64, searchBase ...string) (ad *LDAPClient, err error) {
	protoHostPort := strings.Split(ldapUri, "://")
	if len(protoHostPort) != 2 {
		err = fmt.Errorf("Invalid LDAP URI: %s", ldapUri)
		return
	}

	ad = &LDAPClient{
		URI:          ldapUri,
		UserBindDN:   bindDn,
		UserBindPass: bindPass,
		hostPort:     protoHostPort[1],
		cache:        NewCredentialsCache(cacheTTL),
	}

	if len(searchBase) > 0 {
		ad.DefaultSearchBase = searchBase[0]
	}
	return
}

func (lc *LDAPClient) ConnectAndBind() (err error) {
	if strings.HasPrefix(lc.URI, "ldaps") {
		lc.LdapConn, err = ldap.DialTLS("tcp", lc.hostPort, &tls.Config{InsecureSkipVerify: true})
	} else {
		//no ssl port 389
		lc.LdapConn, err = ldap.Dial("tcp", lc.hostPort)
	}
	if err != nil {
		return
	}
	err = lc.LdapConn.Bind(lc.UserBindDN, lc.UserBindPass)
	return
}

func (lc *LDAPClient) Authenticate(username, password string) (cacheHit bool, err error) {

	if lc.cache.CheckCreds(username, password) {
		cacheHit = true
		return
	}
	cacheHit = false

	var (
		userDn         string
		userLdapClient *LDAPClient
	)

	if err = lc.ConnectAndBind(); err != nil {
		err = fmt.Errorf("Admin bind failed (%s): %s", lc.UserBindDN, err.Error())
		return
	}
	// Close Admin connection
	defer lc.LdapConn.Close()

	if userDn, err = lc.GetUserDN(username); err != nil {
		return
	}

	if userLdapClient, err = NewLDAPClient(lc.URI, userDn, password, 0); err == nil {
		if err = userLdapClient.ConnectAndBind(); err == nil {
			// cache creds
			lc.cache.CacheCreds(NewCredentials(username, userDn, password))
		}
		userLdapClient.LdapConn.Close()
	}

	return
}

// Auth http request
func (lc *LDAPClient) AuthenticateRequest(r *http.Request) (username string, cacheHit bool, err error) {
	var (
		password string
		ok       bool
	)

	if username, password, ok = r.BasicAuth(); !ok {
		err = fmt.Errorf("Unauthorized!")
		return
	}

	if len(username) == 0 || len(password) == 0 {
		err = fmt.Errorf("Username and/or password not provided!")
		return
	}

	if cacheHit, err = lc.Authenticate(username, password); err != nil {
		return
	}
	return
}

func (lc *LDAPClient) GetUserDN(username string) (userDN string, err error) {
	searchReq := &ldap.SearchRequest{
		BaseDN:     lc.DefaultSearchBase,
		Filter:     lc.getUserSearchFilter(username),
		Scope:      ldap.ScopeWholeSubtree,
		Attributes: []string{"distinguishedName"},
	}

	var rslt *ldap.SearchResult
	if rslt, err = lc.LdapConn.Search(searchReq); err != nil {
		return
	}
	if len(rslt.Entries) < 1 {
		err = fmt.Errorf("User not found: %s", username)
		return
	}
	// Double check user dn is actually returned
	if len(rslt.Entries[0].Attributes) < 1 || len(rslt.Entries[0].Attributes[0].Values) < 1 {
		err = fmt.Errorf("User DN attribute not found!")
		return
	}
	userDN = rslt.Entries[0].Attributes[0].Values[0]
	return
}

func (lc *LDAPClient) getUserSearchFilter(samAccountName string) string {
	userFilter := fmt.Sprintf("(&(%s=%s)%s)",
		AD_USER_ATTRIBUTE, ldap.EscapeFilter(samAccountName), AD_USER_FILTER)
	return userFilter
}
