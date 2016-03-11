package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

func (context *Context) AuthenticateMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authorizationHeader := r.Header.Get("authorization")
	requestToken := ""
	if authorizationHeader != "" {
		requestToken = authorizationHeader[len("JWT "):]
	} else {
		signedRequestParam := r.URL.Query().Get("signed_request")
		if signedRequestParam != "" {
			requestToken = signedRequestParam
		} else {
			//return nil, fmt.Errorf("Request is missing an authorization header")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	verifiedToken, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Debugf("invalid token: %s", token)
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		issuer, ok := token.Claims["iss"].(string)
		if !ok {
			return nil, fmt.Errorf("JWT claim did not contain the issuer (iss) claim")
		}

		group, err := GetGroupByOauthId(context, issuer)

		if err != nil {
			log.Debugf("Couldn't find group with oauthId-%s", issuer)
			return nil, fmt.Errorf("Request can't be verified without a valid OAuth secret")
		}

		if token.Header["alg"] != "HS256" {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])

		}

		return []byte(group.oauthSecret), nil
	})

	if err == nil && verifiedToken.Valid {
		context.token = verifiedToken
		next(w, r)
	} else {
		log.Errorf("Authentication error %s", err)
		w.WriteHeader(http.StatusUnauthorized)
	}
}
