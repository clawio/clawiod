// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package pat defines the Pat interface and provides an implementation.
package pat

import (
	"errors"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/github.com/dgrijalva/jwt-go"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"net/http"
	"time"
)

// Pat is the interface that authentication pats must implement.
type Pat interface {
	AddAuthType(authType auth.AuthType) error
	Authenticate(req *http.Request, authTypeID string) (auth.Identity, error)
	BasicAuthenticate(username, password string) (auth.Identity, error)
	CreateToken(identity auth.Identity) (string, error)
	ValidateRequest(r *http.Request) (auth.Identity, error)
	ValidateRequestHandler(ctx context.Context,
		w http.ResponseWriter, r *http.Request, sendBasicChallenge bool,
		next func(ctx context.Context, w http.ResponseWriter, r *http.Request))
}

// pat dispatchs authentication request to the proper backend.
type pat struct {
	auths map[string]auth.AuthType
	config.Config
	logger.Logger
}

// New creates an pat object or returns an error
func New(cfg config.Config, log logger.Logger) Pat {
	return &pat{auths: make(map[string]auth.AuthType), Config: cfg, Logger: log}
}

// AddAuthType register an authentication providers to be
// used for authenticate requests.
func (d *pat) AddAuthType(authType auth.AuthType) error {
	if _, ok := d.auths[authType.ID()]; ok {
		return fmt.Errorf("pat: authtype:%s already registered", authType.ID())
	}
	d.auths[authType.ID()] = authType
	return nil
}

// Authenticate authenticates an user.
func (d *pat) Authenticate(req *http.Request,
	authTypeID string) (auth.Identity, error) {

	// the authentication request has been made
	// specifically for an authentication backend.
	if authTypeID != "" {
		a, ok := d.auths[authTypeID]
		if ok {
			identity, err := a.Authenticate(req)
			if err != nil {
				return nil, err
			}
			return identity, nil
		}
		return nil, &auth.IdentityNotFoundError{AuthTypeID: a.ID()}
	}

	// if the auth backend with the authTypeID passed is not
	// found we try all the auth providers.
	// This is needed because with Basic Auth we cannot
	// send the auth provider ID.
	for _, a := range d.auths {
		identity, err := a.Authenticate(req)
		if err == nil {
			return identity, nil
		}
	}

	// we couldn´t find any auth provider that authenticated this user
	return nil, &auth.IdentityNotFoundError{AuthTypeID: "all"}
}

// ValidateRequest authenticates a HTTP request.
// It returns an Identity object or an error.
//
//    This method DOES NOT create an HTTP response with 401
//    if the authentication fails. To handle HTTP responses
//    you must do it yourself or use the ValidateRequestHandler mehtod.
//    The following mechanisms are used:
//    1. JWT token as query parameter in the URL.
//    2. JWT token in HTTP Header.
//    3. HTTP Basic Authentication without digest
func (d *pat) ValidateRequest(r *http.Request) (auth.Identity, error) {
	// 1. JWT authentication token in query parameter.
	param := r.URL.Query().Get(d.GetDirectives().AuthTokenQueryParamName)
	if param != "" {
		token, err := jwt.Parse(param,
			func(token *jwt.Token) (key interface{}, err error) {
				return []byte(d.GetDirectives().TokenSecret), nil
			})

		if err != nil {
			msg := "apat: failed parsing auth query param. err: %s"
			return nil, fmt.Errorf(msg, err.Error())
		}

		return d.getIdentityFromToken(token)
	}

	// 2. JWT authentication token in HTTP header.
	authHeader := r.Header.Get(d.GetDirectives().AuthTokenHeaderName)
	if authHeader != "" {
		token, err := jwt.Parse(authHeader,
			func(token *jwt.Token) (key interface{}, err error) {
				return []byte(d.GetDirectives().TokenSecret), nil
			})

		if err != nil {
			msg := "apat: failed parsing auth header. err: %s"
			return nil, fmt.Errorf(msg, err.Error())
		}

		return d.getIdentityFromToken(token)
	}

	// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
	username, password, ok := r.BasicAuth()
	if ok {
		identity, err := d.BasicAuthenticate(username, password)
		if err != nil {
			return nil, err
		}
		return identity, nil
	}

	return nil, errors.New("apat: no credentials in req")
}

func (d *pat) BasicAuthenticate(username,
	password string) (auth.Identity, error) {

	for _, a := range d.auths {
		if a.Capabilities().BasicAuth() {
			identity, err := a.BasicAuthenticate(username, password)
			if err == nil {
				return identity, nil
			}
		}
	}
	return nil, &auth.IdentityNotFoundError{AuthTypeID: "all"}
}

// CreateToken creates an JWT authentication token from an Identity.
// It returns the JWT token or an error.
func (d *pat) CreateToken(identity auth.Identity) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(d.GetDirectives().TokenCipherSuite))
	token.Claims["pid"] = identity.PID()
	token.Claims["idp"] = identity.IDP()
	token.Claims["displayname"] = identity.DisplayName()
	token.Claims["email"] = identity.Email()
	token.Claims["authid"] = identity.AuthTypeID()
	token.Claims["iss"] = d.GetDirectives().TokenISS
	token.Claims["exp"] = time.Now().Add(time.Second *
		time.Duration(d.GetDirectives().TokenExpirationTime)).UnixNano()

	tokenStr, err := token.SignedString([]byte(d.GetDirectives().TokenSecret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// ValidateRequestHandler is an HTTP middleware that besides authenticating
// the request like the ValidateRequest method
// it does the following:
//
// 1. Return 401 (Unauthorized) if the authentication fails.
//
// 2. Save the Identity object in the request context and call the next handler
// if the authentication is successful.
func (d *pat) ValidateRequestHandler(ctx context.Context, w http.ResponseWriter,
	r *http.Request, sendBasicChallenge bool,
	next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {

	identity, err := d.ValidateRequest(r)
	if err != nil {
		d.Err("apat: " + err.Error())
		if sendBasicChallenge {
			w.Header().Set(
				"WWW-Authenticate", "Basic Realm='ClawIO credentials'",
			)
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)

		return
	}
	ctx = context.WithValue(ctx, "idt", identity)
	next(ctx, w, r)
}

func (d *pat) getIdentityFromToken(token *jwt.Token) (auth.Identity, error) {
	identity := &identity{}

	pidString, ok := token.Claims["pid"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of pid:%s",
			fmt.Sprintln(token.Claims["pid"]))
	}
	idpString, ok := token.Claims["idp"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of idp:%s",
			fmt.Sprintln(token.Claims["idp"]))
	}
	displaynameString, ok := token.Claims["displayname"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of displayname:%s",
			fmt.Sprintln(token.Claims["displayname"]))
	}
	emailString, ok := token.Claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of email:%s",
			fmt.Sprintln(token.Claims["email"]))
	}
	authidString, ok := token.Claims["authid"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of authid:%s",
			fmt.Sprintln(token.Claims["authid"]))
	}

	identity.pid = pidString
	identity.idp = idpString
	identity.displayName = displaynameString
	identity.email = emailString
	identity.authTypeID = authidString
	identity.extra = token.Claims["extra"]

	return identity, nil
}

type identity struct {
	pid         string
	idp         string
	authTypeID  string
	email       string
	displayName string
	extra       interface{}
}

func (i *identity) PID() string         { return i.pid }
func (i *identity) IDP() string         { return i.idp }
func (i *identity) AuthTypeID() string  { return i.idp }
func (i *identity) Email() string       { return i.email }
func (i *identity) DisplayName() string { return i.displayName }
func (i *identity) Extra() interface{}  { return i.extra }
