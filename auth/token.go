package auth

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const DEFAULT_TOKEN_ISS = "vindalu" // token issuer

func GetNewToken(username string, ttl int64) (token *jwt.Token) {

	timestamp := time.Now()

	token = jwt.New(jwt.SigningMethodHS256)
	token.Claims["iss"] = DEFAULT_TOKEN_ISS
	token.Claims["sub"] = username
	token.Claims["iat"] = timestamp.Unix()
	token.Claims["exp"] = timestamp.Add(time.Second * time.Duration(ttl)).Unix()

	return
}

func GetNewSignedToken(username string, ttl int64, signkey []byte) (string, error) {
	return GetNewToken(username, ttl).SignedString(signkey)
}

func GetTokenFromString(tokenStr string, signkey []byte) (token *jwt.Token, err error) {
	token, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signature method: %s!", t.Method)
		}
		return signkey, nil
	})

	err = validateToken(token)

	return
}

func GetTokenFromRequest(r *http.Request, signkey []byte) (token *jwt.Token, err error) {

	if token, err = jwt.ParseFromRequest(r, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid signature method: %s!", t.Method)
		}
		return signkey, nil
	}); err != nil {
		return
	}

	err = validateToken(token)

	return
}

func validateToken(token *jwt.Token) (err error) {
	if !token.Valid {
		fval, _ := token.Claims["exp"].(int64)

		if fval < time.Now().Unix() {
			err = fmt.Errorf("Token expired!")
		} else {
			err = fmt.Errorf("Token invalid!")
		}
	}
	return
}
