package auth

import (
	"net/http"
)

const (
	AD_USER_FILTER        = "(objectClass=user)"
	AD_USER_ATTRIBUTE     = "sAMAccountName"
	AD_GROUP_ATTRIBUTE    = "memberOf"
	AD_GROUP_OBJECT_CLASS = "group"
)

type LocalAuthGroups map[string][]string

type IAuthenticator interface {
	AuthenticateRequest(r *http.Request) (username string, cacheHit bool, err error)
}
