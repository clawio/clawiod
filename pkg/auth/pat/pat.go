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

// Pat is the interface that identity manager dispatchers must implement.
type Pat interface {
	AddIDM(idm idm.IDM) error
	Authenticate(ctx context.Context, req *http.Request, idmID string) (*idm.Identity, error)
	BasicAuthenticate(ctx context.Context, username, password string) (*idm.Identity, error)
	CreateToken(ctx context.Context, idt *idm.Identity) (string, error)
	ValidateRequest(ctx context.Context, r *http.Request) (*idm.Identity, error)
	ValidateRequestHandler(ctx context.Context,
		w http.ResponseWriter, r *http.Request, sendBasicChallenge bool,
		next func(ctx context.Context, w http.ResponseWriter, r *http.Request))
}

// pat dispatchs identity manager request to the proper backend.
type pat struct {
	idms map[string]idm.IDM
	cfg  config.Config
}

// New creates a pat object or returns an error
func New(cfg config.Config) Pat {
	return &pat{idms: make(map[string]idm.IDM), cfg: cfg}
}

// AddIDM registers an identity manager.
func (d *pat) AddIDM(idm idm.IDM) error {
	if _, ok := d.idms[idm.ID()]; ok {
		return fmt.Errorf("idmpat: idmtype:%s already registered", idm.ID())
	}
	d.idms[idm.ID()] = idm
	return nil
}

// Authenticate authenticates an user.
func (d *pat) Authenticate(ctx context.Context, req *http.Request,
	idmID string) (*idm.Identity, error) {

	// the request has been made to a concrete identity manager.
	if idmID != "" {
		a, ok := d.idms[idmID]
		if ok {
			idt, err := a.Authenticate(ctx, req)
			if err != nil {
				return nil, err
			}
			return idt, nil
		}
		return nil, &idm.IdentityNotFoundError{IDMID: a.ID()}
	}

	// if the idm backend with the idmID passed is not
	// found we try all the idm.
	// This is needed because with Basic Auth we cannot
	// send the idm provider ID.
	for _, a := range d.idms {
		identity, err := a.Authenticate(ctx, req)
		if err == nil {
			return identity, nil
		}
	}

	// we couldn´t find any idm provider that idmenticated this user
	return nil, &idm.IdentityNotFoundError{IDMID: "all"}
}

// ValidateRequest authenticates a HTTP request.
// It returns an Identity object or an error.
//
//    This method DOES NOT create an HTTP response with 401
//    if the identity manager fails. To handle HTTP responses
//    you must do it yourself or use the ValidateRequestHandler mehtod.
//    The following mechanisms are used:
//    1. JWT token as query parameter in the URL.
//    2. JWT token in HTTP Header.
//    3. HTTP Basic Authentication.
func (d *pat) ValidateRequest(ctx context.Context, r *http.Request) (*idm.Identity, error) {
	// 1. JWT identity manager token in query parameter.
	param := r.URL.Query().Get(d.cfg.GetDirectives().AuthTokenQueryParamName)
	if param != "" {
		token, err := jwt.Parse(param,
			func(token *jwt.Token) (key interface{}, err error) {
				return []byte(d.cfg.GetDirectives().TokenSecret), nil
			})

		if err != nil {
			msg := "idmpat: failed parsing idm query param. err: %s"
			return nil, fmt.Errorf(msg, err.Error())
		}

		return d.getIdentityFromToken(token)
	}

	// 2. JWT identity manager token in HTTP header.
	idmHeader := r.Header.Get(d.cfg.GetDirectives().AuthTokenHeaderName)
	if idmHeader != "" {
		token, err := jwt.Parse(idmHeader,
			func(token *jwt.Token) (key interface{}, err error) {
				return []byte(d.cfg.GetDirectives().TokenSecret), nil
			})

		if err != nil {
			msg := "idmpat: failed parsing auth header. err: %s"
			return nil, fmt.Errorf(msg, err.Error())
		}

		return d.getIdentityFromToken(token)
	}

	// 3. HTTP Basic Authentication without digest (Plain Basic Auth).
	username, password, ok := r.BasicAuth()
	if ok {
		identity, err := d.BasicAuthenticate(ctx, username, password)
		if err != nil {
			return nil, err
		}
		return identity, nil
	}

	return nil, errors.New("idmpat: no credentials in req")
}

func (d *pat) BasicAuthenticate(ctx context.Context, username,
	password string) (*idm.Identity, error) {

	for _, a := range d.idms {
		if a.Capabilities(ctx).BasicAuth {
			identity, err := a.BasicAuthenticate(ctx, username, password)
			if err == nil {
				return identity, nil
			}
		}
	}
	return nil, &idm.IdentityNotFoundError{IDMID: "all"}
}

// CreateToken creates an JWT token from Identity.
// It returns the JWT token or an error.
func (d *pat) CreateToken(ctx context.Context, idt *idm.Identity) (string, error) {
	token := jwt.New(jwt.GetSigningMethod(d.cfg.GetDirectives().TokenCipherSuite))
	token.Claims["pid"] = idt.PID
	token.Claims["idp"] = idt.IDP
	token.Claims["displayname"] = idt.DisplayName
	token.Claims["email"] = idt.Email
	token.Claims["idmid"] = idt.IDMID
	token.Claims["iss"] = d.cfg.GetDirectives().TokenISS
	token.Claims["exp"] = time.Now().Add(time.Second *
		time.Duration(d.cfg.GetDirectives().TokenExpirationTime)).UnixNano()

	tokenStr, err := token.SignedString([]byte(d.cfg.GetDirectives().TokenSecret))
	if err != nil {
		return "", err
	}
	return tokenStr, nil
}

// ValidateRequestHandler is an HTTP middleware that besides authenticating
// the request like the ValidateRequest method it does the following:
//
// 1. Return 401 (Unauthorized) if the identity manager fails authentication.
//
// 2. Save the Identity object in the request context and call the next handler
// if the identity manager is successful.
func (d *pat) ValidateRequestHandler(ctx context.Context, w http.ResponseWriter,
	r *http.Request, sendBasicChallenge bool,
	next func(ctx context.Context, w http.ResponseWriter, r *http.Request)) {

	log := logger.MustFromContext(ctx)

	idt, err := d.ValidateRequest(ctx, r)
	if err != nil {
		log.Err("idmpat: " + err.Error())
		if sendBasicChallenge {
			w.Header().Set(
				"WWW-Authenticate", "Basic Realm='ClawIO credentials'",
			)
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)

		return
	}
	ctx = idm.NewContext(ctx, idt)
	next(ctx, w, r)
}

func (d *pat) getIdentityFromToken(token *jwt.Token) (*idm.Identity, error) {
	identity := &idm.Identity{}

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
	idmidString, ok := token.Claims["idmid"].(string)
	if !ok {
		return nil, fmt.Errorf("failed cast to string of idmid:%s",
			fmt.Sprintln(token.Claims["idmid"]))
	}

	identity.PID = pidString
	identity.IDP = idpString
	identity.DisplayName = displaynameString
	identity.Email = emailString
	identity.IDMID = idmidString
	identity.Extra = token.Claims["extra"]

	return identity, nil
}

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// patKey is the context key for the dispatcher.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const patKey key = 0

// NewContext returns a new Context carrying an IDM pat.
func NewContext(ctx context.Context, p Pat) context.Context {
	return context.WithValue(ctx, patKey, p)
}

// FromContext extracts the IDM pat from ctx, if present.
func FromContext(ctx context.Context) (Pat, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	p, ok := ctx.Value(patKey).(Pat)
	return p, ok
}

// MustFromContext extracts the IDM pat from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) Pat {
	pat, ok := ctx.Value(patKey).(Pat)
	if !ok {
		panic("idm pat is not registered")
	}
	return pat
}
