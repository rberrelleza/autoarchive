package web

import (
	"fmt"
	"net/http"

	"bitbucket.org/rbergman/go-hipchat-connect/tenant"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

const tenantKey string = "hcc:tenant"

// GetTenant returns the authenticated tenant from the current request, or an
// error otherwise.
func GetTenant(r *http.Request) (*tenant.Tenant, error) {
	if rv := context.Get(r, tenantKey); rv != nil {
		t, ok := rv.(*tenant.Tenant)
		if !ok {
			return nil, fmt.Errorf("Tenant found but not of correct type")
		}
		return t, nil
	}

	return nil, fmt.Errorf("No tenant found in current request")
}

func setTenant(r *http.Request, val *tenant.Tenant) {
	context.Set(r, tenantKey, val)
}

// Authenticate is a Negroni middleware for verifying JWT authentication. When
// successful authentication is performed, it sets the currently authenticated
// tenant on the request, which can be retrieved using the `GetTenant` helper
// defined in this package. It supports both query parameter
// (`signed_request=<token>`) and header (`Authorization: JWT <token>`)
// verification, in that order.
type Authenticate struct {
	Server *Server
}

// NewAuthenticate returns a new Authenticate instance.
func NewAuthenticate(s *Server) *Authenticate {
	return &Authenticate{s}
}

func (a *Authenticate) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	a.Server.Log.Debugf("Authenticating request")
	authorizationHeader := r.Header.Get("authorization")
	requestToken := ""
	if authorizationHeader != "" {
		requestToken = authorizationHeader[len("JWT "):]
	} else {
		signedRequestParam := r.URL.Query().Get("signed_request")
		if signedRequestParam != "" {
			requestToken = signedRequestParam
		} else {
			a.Server.Log.Debugf("Authentication parameter was missing")
			err := fmt.Errorf("JWT token missing from request")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	var tenant *tenant.Tenant

	verifiedToken, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("Unexpected signing method: %s", token.Header["alg"])
			return nil, err
		}

		issuer, ok := token.Claims["iss"].(string)
		if !ok {
			err := fmt.Errorf("JWT claim did not contain the issuer (iss) claim")
			return nil, err
		}

		tenants := a.Server.NewTenants()
		t, err := tenants.Get(issuer)

		if err != nil {
			a.Server.Log.Debugf("Couldn't find group with oauthId-%s", issuer)
			err = fmt.Errorf("Request can't be verified without a valid OAuth secret")
			return nil, err
		}

		if token.Header["alg"] != "HS256" {
			err = fmt.Errorf("Unexpected signing method: %s", token.Header["alg"])
			return nil, err
		}

		tenant = t

		return []byte(tenant.Secret), nil
	})

	if err == nil && verifiedToken.Valid {
		a.Server.Log.Debugf("Authenticated request successfully, calling the next handler in the pipeline")
		a.Server.Log.Debugf("Tenant %v", tenant)
		setTenant(r, tenant)
		next(w, r)
	} else {
		a.Server.Log.Debugf("Error authenticating request")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
}
