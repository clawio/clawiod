// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package dispatcher defines the authentication multiplexer to authenticate requests against
// the registered authentication providers.
package dispatcher

import (
	"errors"
	"fmt"
	"github.com/clawio/clawiod/lib/config"
	"net/http"
	"time"

	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/logger"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
)

// Dispatcher is the interface that auth providers must implement.
type Dispatcher interface {
	AddAuth(ap auth.Auth) error
	Authenticate(username, password, id string, extra interface{}) (*auth.Identity, error)
	AuthenticateRequest(r *http.Request) (*auth.Identity, error)
	AuthenticateRequestWithMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, next func(ctx context.Context, w http.ResponseWriter, r *http.Request))
	CreateAuthTokenFromIdentity(identity *auth.Identity) (string, error)
}

// dispatcher is the multiplexer responsible for routing authentication to an specific
// authentication provider.
// It keeps a map with all the authentication providers registered.
type dispatcher struct {
	auths map[string]auth.Auth
	cfg   *config.Config
	log   logger.Logger
}

// New creates an dispatcher object or returns an error
func New(cfg *config.Config, log logger.Logger) Dispatcher {
	m := dispatcher{}
	m.cfg = cfg
	m.log = log
	m.auths = make(map[string]auth.Auth)

	return &m
}

// AddAuth register an authentication providers to be used for authenticate requests.
func (d *dispatcher) AddAuth(ap auth.Auth) error {
	if _, ok := d.auths[ap.GetID()]; ok {
		return fmt.Errorf("auth '%s' already registered", ap.GetID())
	}
	d.auths[ap.GetID()] = ap
	return nil
}

// Authenticate authenticates a user with username and password credentials.
// The id parameter is the authentication provider id.
func (d *dispatcher) Authenticate(username, password, id string, extra interface{}) (*auth.Identity, error) {
	// the authentication request has been made specifically for an authentication provider.
	if id != "" {
		a, ok := d.auths[id]
		// if an auth provider with the id passed is found we just use this auth provider.
		if ok {
			identity, err := a.Authenticate(username, password, extra)
			if err != nil {
				return nil, err
			}
			return identity, nil
		}
		return nil, &auth.IdentityNotFoundError{ID: username, AuthID: id}
	}

	// if the auth provider with the id passed is not found we try all the auth providers.
	// This is needed because with Basic Auth we cannot send the auth provider ID.
	for _, a := range d.auths {
		if a.GetID() != id {
			aRes, _ := a.Authenticate(username, password, extra)
			if aRes != nil {
				return aRes, nil
			}
		}
	}

	// we couldn´t find any auth provider that authenticated this user
	return nil, &auth.IdentityNotFoundError{ID: username, AuthID: "all"}
}

// AuthenticateRequest authenticates a HTTP request.
//
// It returns an AuthenticationResource object or an error.
//
// This method DOES NOT create an HTTP response with 401 if the authentication fails. To handle HTTP responses
// you must do it yourself or use the AuthenticateRequestWithMiddleware mehtod.
//
// The following mechanisms are used in the order described to authenticate the request.
//
// 1. JWT authentication token as query parameter in the URL. The parameter name is auth-key.
//
// 2. JWT authentication token in the HTTP Header called X-Auth-Key.
//
// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
//
// More authentication methods wil be used in the future like Kerberos access tokens.
func (d *dispatcher) AuthenticateRequest(r *http.Request) (*auth.Identity, error) {
	// 1. JWT authentication token in query parameter.
	authQueryParam := r.URL.Query().Get(d.cfg.GetDirectives().AuthTokenQueryParamName)
	if authQueryParam != "" {
		token, err := jwt.Parse(authQueryParam, func(token *jwt.Token) (key interface{}, err error) {
			return []byte(d.cfg.GetDirectives().TokenSecret), nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed parsing auth query param because: %s", err.Error())
		}
		identity := &auth.Identity{}
		identity.ID = token.Claims["id"].(string)
		identity.DisplayName = token.Claims["display_name"].(string)
		identity.Email = token.Claims["email"].(string)
		identity.AuthID = token.Claims["auth_id"].(string)

		return identity, nil
	}

	// 2. JWT authentication token in HTTP header.
	authHeader := r.Header.Get(d.cfg.GetDirectives().AuthTokenHeaderName)
	if authHeader != "" {
		token, err := jwt.Parse(authHeader, func(token *jwt.Token) (key interface{}, err error) {
			return []byte(d.cfg.GetDirectives().TokenSecret), nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed parsing auth header because: %s", err.Error())
		}
		// TODO: be sure we handle proper conversion
		identity := &auth.Identity{}
		identity.ID = token.Claims["id"].(string)
		identity.DisplayName = token.Claims["display_name"].(string)
		identity.Email = token.Claims["email"].(string)
		identity.AuthID = token.Claims["auth_id"].(string)
		identity.Extra = token.Claims["extra"]

		return identity, nil
	}

	// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
	username, password, ok := r.BasicAuth()
	if ok {
		identity, err := d.Authenticate(username, password, "", nil)
		if err != nil {
			return nil, err
		}
		if err == nil {
			return identity, nil
		}
	}

	return nil, errors.New("no auth credentials found in the request")
}

// CreateAuthTokenFromIdentity creates an JWT authentication token from an AuthenticationResource object.
// It returns the JWT token or an error.
func (d *dispatcher) CreateAuthTokenFromIdentity(identity *auth.Identity) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(d.cfg.GetDirectives().TokenCipherSuite))
	token.Claims["iss"] = d.cfg.GetDirectives().TokenISS
	token.Claims["exp"] = time.Now().Add(time.Minute * 480).Unix() // we need to use cfg.TokenExpirationTime
	token.Claims["id"] = identity.ID
	token.Claims["display_name"] = identity.DisplayName
	token.Claims["email"] = identity.Email
	token.Claims["auth_id"] = identity.AuthID

	tokenString, err := token.SignedString([]byte(d.cfg.GetDirectives().TokenSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// AuthenticateRequestWithMiddleware is an HTTP middleware that besides authenticating the request like the AuthenticateRequest method
// it does the following:
//
// 1. Return 401 (Unauthorized) if the authentication fails.
//
// 2. Save the Identity object in the request context and call the next handler if the authentication is successful.
func (d *dispatcher) AuthenticateRequestWithMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {
	identity, err := d.AuthenticateRequest(r)
	if err != nil {
		d.log.Warningf("Authentication of request failed: %+v", map[string]interface{}{"err": err})
		w.Header().Set("WWW-Authenticate", "Basic Realm='ClawIO credentials'")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	d.log.Infof("Authentication of request successful: %+v", map[string]interface{}{"username": identity.ID, "auth_id": identity.AuthID})
	ctx = context.WithValue(ctx, "identity", identity)
	next(ctx, w, r)
}
