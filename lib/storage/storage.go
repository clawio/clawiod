// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package storage defines the interface that storage providers must implement and
// defines the metadata resource.
package storage

import (
	"github.com/clawio/clawiod/lib/auth"
	"io"
	"net/url"
)

// Storage is the interface that all the storage providers must implement
// to be used by the storage multiplexer.
// An storage provider is defined by an id called Scheme.
//
// A resource is uniquely identified by a URI http://en.wikipedia.org/wiki/Uniform_resource_identifier
type Storage interface {

	// GetScheme returns the scheme/id of this storage.
	GetScheme() string

	// CreateUserHome creates the user home directory in the storage.
	CreateUserHome(authRes *auth.Identity) error

	// IsUserHomeCreated checks if the user home directory has been created or not.
	IsUserHomeCreated(authRes *auth.Identity) (bool, error)

	// PutFile puts a file into the storage defined by the uri.
	// If the error returned is not in this list it must be considered
	// an unexcepted error
	//   NotExistError
	//   ExistError
	PutFile(authRes *auth.Identity, uri *url.URL, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error

	// GetFile gets a file from the storage defined by the uri or by the resourceID.
	GetFile(authRes *auth.Identity, uri *url.URL) (io.Reader, error)

	// Stat returns metadata information about the resources and its children.
	Stat(authRes *auth.Identity, uri *url.URL, children bool) (*MetaData, error)

	// Remove removes a resource from the storage defined by the uri.
	Remove(authRes *auth.Identity, uri *url.URL, recursive bool) error

	// CreateCol creates a collection in the storage defined by the uri.
	CreateCol(authRes *auth.Identity, uri *url.URL, recursive bool) error

	// Copy copies a resource from one uri to another.
	// If uris belong to different storages this is a cross-storage copy.
	Copy(authRes *auth.Identity, fromURI, toURI *url.URL) error

	// Rename renames/move a resource from one uri to another.
	// If uris belong to different storages this is a cross-storage rename.
	Rename(authRes *auth.Identity, fromURI, toURI *url.URL) error

	// GetCapabilities returns the capabilities of this storage.
	GetCapabilities() *Capabilities

	GetSupportedChecksumTypes() []string
}
