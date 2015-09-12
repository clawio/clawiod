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
	"github.com/clawio/clawiod/pkg/config"
	"net/http"
	"time"

	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/logger"

	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/dgrijalva/jwt-go"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
)

// Dispatcher is the interface that authentication dispatchers must implement.
type Dispatcher interface {
	AddAuthenticationstrategy(authStrategy auth.AuthenticationStrategy) error
	AuthenticateRequest(r *http.Request) (*auth.Identity, error)
	AuthenticateRequestWithMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, sendBasicChallenge bool, next func(ctx context.Context, w http.ResponseWriter, r *http.Request))
	CreateAuthTokenFromIdentity(identity *auth.Identity) (string, error)
	DispatchAuthenticate(eppn, password, idp string, extra interface{}, authID string) (*auth.Identity, error)
}

// dispatcher dispatchs authentication request to the proper backend.
type dispatcher struct {
	auths map[string]auth.AuthenticationStrategy
	cfg   *config.Config
	log   logger.Logger
}

// New creates an dispatcher object or returns an error
func New(cfg *config.Config, log logger.Logger) Dispatcher {
	m := dispatcher{}
	m.cfg = cfg
	m.log = log
	m.auths = make(map[string]auth.AuthenticationStrategy)
	return &m
}

// AddAuthenticationstrategy register an authentication providers to be used for authenticate requests.
func (d *dispatcher) AddAuthenticationstrategy(authStrategy auth.AuthenticationStrategy) error {
	if _, ok := d.auths[authStrategy.GetID()]; ok {
		return fmt.Errorf("authentication backend %s already registered", authStrategy.GetID())
	}
	d.auths[authStrategy.GetID()] = authStrategy
	return nil
}

// DispatchAuthenticate authenticates a user with eppn and password credentials.
// The id parameter is the authentication provider id.
func (d *dispatcher) DispatchAuthenticate(eppn, password, idp string, extra interface{}, authID string) (*auth.Identity, error) {
	// the authentication request has been made specifically for an authentication backend.
	if authID != "" {
		a, ok := d.auths[authID]
		// if an auth backend with the authID passed is found we just use this auth provauthIDer.
		if ok {
			identity, err := a.Authenticate(eppn, password, idp, extra)
			if err != nil {
				return nil, err
			}
			return identity, nil
		}
		return nil, &auth.IdentityNotFoundError{EPPN: eppn, IdP: idp, AuthID: authID}
	}

	// if the auth backend with the authID passed is not found we try all the auth providers.
	// This is needed because with Basic Auth we cannot send the auth provider ID.
	for _, a := range d.auths {
		identity, err := a.Authenticate(eppn, password, idp, extra)
		if err == nil {
			return identity, nil
		}
	}

	// we couldn´t find any auth provider that authenticated this user
	return nil, &auth.IdentityNotFoundError{EPPN: eppn, AuthID: "all"}
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

		eppnString, ok := token.Claims["eppn"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of eppn:%s", fmt.Sprintln(token.Claims["eppn"]))
		}
		idpString, ok := token.Claims["idp"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of idp:%s", fmt.Sprintln(token.Claims["idp"]))
		}
		displaynameString, ok := token.Claims["displayname"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of displayname:%s", fmt.Sprintln(token.Claims["displayname"]))
		}
		emailString, ok := token.Claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of email:%s", fmt.Sprintln(token.Claims["email"]))
		}
		authidString, ok := token.Claims["authid"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of authid:%s", fmt.Sprintln(token.Claims["authid"]))
		}

		identity.EPPN = eppnString
		identity.IdP = idpString
		identity.DisplayName = displaynameString
		identity.Email = emailString
		identity.AuthID = authidString
		identity.Extra = token.Claims["extra"]

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
		identity := &auth.Identity{}

		eppnString, ok := token.Claims["eppn"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of eppn:%s", fmt.Sprintln(token.Claims["eppn"]))
		}
		idpString, ok := token.Claims["idp"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of idp:%s", fmt.Sprintln(token.Claims["idp"]))
		}
		displaynameString, ok := token.Claims["displayname"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of displayname:%s", fmt.Sprintln(token.Claims["displayname"]))
		}
		emailString, ok := token.Claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of email:%s", fmt.Sprintln(token.Claims["email"]))
		}
		authidString, ok := token.Claims["authid"].(string)
		if !ok {
			return nil, fmt.Errorf("failed cast to string of authid:%s", fmt.Sprintln(token.Claims["authid"]))
		}

		identity.EPPN = eppnString
		identity.IdP = idpString
		identity.DisplayName = displaynameString
		identity.Email = emailString
		identity.AuthID = authidString
		identity.Extra = token.Claims["extra"]

		return identity, nil
	}

	// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
	eppn, password, ok := r.BasicAuth()
	if ok {
		identity, err := d.DispatchAuthenticate(eppn, password, "", nil, "")
		if err != nil {
			return nil, err
		}
		return identity, nil
	}

	return nil, errors.New("no auth credentials found in the request")
}

// CreateAuthTokenFromIdentity creates an JWT authentication token from an AuthenticationResource object.
// It returns the JWT token or an error.
func (d *dispatcher) CreateAuthTokenFromIdentity(identity *auth.Identity) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(d.cfg.GetDirectives().TokenCipherSuite))
	token.Claims["eppn"] = identity.EPPN
	token.Claims["idp"] = identity.IdP
	token.Claims["displayname"] = identity.DisplayName
	token.Claims["email"] = identity.Email
	token.Claims["authid"] = identity.AuthID
	token.Claims["iss"] = d.cfg.GetDirectives().TokenISS
	token.Claims["exp"] = time.Now().Add(time.Second * time.Duration(d.cfg.GetDirectives().TokenExpirationTime)).Unix()

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
func (d *dispatcher) AuthenticateRequestWithMiddleware(ctx context.Context, w http.ResponseWriter, r *http.Request, sendBasicChallenge bool, next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {
	identity, err := d.AuthenticateRequest(r)
	if err != nil {
		if sendBasicChallenge {
			w.Header().Set("WWW-Authenticate", "Basic Realm='ClawIO credentials'")
		}
		//d.log.Warningf("Authentication of request failed: %+v", map[string]interface{}{"err": err})
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	//d.log.Infof("Authentication of request successful: %+v", map[string]interface{}{"eppn": identity.EPPN, "idp": identity.IdP, "authid": identity.AuthID})
	ctx = context.WithValue(ctx, "identity", identity)
	next(ctx, w, r)
}
