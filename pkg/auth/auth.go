// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package idm defines the interface that identity managers must
// implement and defines the identity resource.
package idm

import (
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"net/http"
)

// IDM is the interface that all the identity managers must implement
// to be used by the identity manager dispatcher.
// An identity manager backend has a unique identifier, the ID.
type IDM interface {
	ID() string
	Authenticate(ctx context.Context, req *http.Request) (*Identity, error)
	BasicAuthenticate(ctx context.Context, username, password string) (*Identity, error)
	Capabilities(ctx context.Context) *Capabilities
}

// Identity represents the details of an idmenticated user.
type Identity struct {
	PID         string      `json:"pid"`
	IDP         string      `json:"idp"`
	IDMID       string      `json:"idmid"`
	Email       string      `json:"email"`
	DisplayName string      `json:"displayname"`
	Extra       interface{} `json:"extra"`
}

func (i *Identity) String() string {
	return fmt.Sprintf("identity(%+v)", *i)
}

// IdentityNotFoundError represents a missing user in
// the identity manager backend.
type IdentityNotFoundError struct {
	PID   string
	IDP   string
	IDMID string
}

func (e *IdentityNotFoundError) Error() string {
	return fmt.Sprintf("identity (eppn:%s idp:%s idmid:%s) not found",
		e.PID, e.IDP, e.IDMID)
}

type Capabilities struct {
	BasicAuth bool `json:"basicidm"`
}

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// idtKey is the context key for a identity.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const idtKey key = 0

// NewContext returns a new Context carrying an Identity pat.
func NewContext(ctx context.Context, idt *Identity) context.Context {
	return context.WithValue(ctx, idtKey, idt)
}

// FromContext extracts the Identity pat from ctx, if present.
func FromContext(ctx context.Context) (*Identity, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	p, ok := ctx.Value(idtKey).(*Identity)
	return p, ok
}

// MustFromContext extracts the identity from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) *Identity {
	idt, ok := ctx.Value(idtKey).(*Identity)
	if !ok {
		panic("identity is not registered")
	}
	return idt
}
