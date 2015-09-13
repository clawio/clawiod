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
	"github.com/clawio/clawiod/pkg/auth"
	"io"
)

// Storage is the interface that all the storage backends must implement
// to be used by the storage multiplexer.
type Storage interface {

	// GetCapabilities returns the capabilities of this storage.
	GetCapabilities(identity *auth.Identity) *Capabilities

	// GetStoragePrefix returns the prefix of this storage.
	GetStoragePrefix() string

	// CreateUserHomeDirectory creates the user home directory in the storage.
	CreateUserHomeDirectory(identity *auth.Identity) error

	// PutObject puts an object into the storage defined by resourcePath.
	PutObject(identity *auth.Identity, resourcePath string, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error

	// StartChunkedUpload starts a transaction for putting an object in chunks.
	StartChunkedUpload() (string, error)

	// PutChunkedObject uploads the chunk defined by start and size of an object.
	PutChunkedObject(identity *auth.Identity, r io.Reader, size int64, start int64, chunkID string) error

	// CommitChunkedUpload commits the transaction
	CommitChunkedUpload(chunkID string, verifyChecksum bool, checksum, checksumType string) error

	// GetObject gets an object from the storage defined by the uri or by the resourceID.
	GetObject(identity *auth.Identity, resourcePath string) (io.Reader, error)

	// Stat returns metadata information about the resources and its children.
	Stat(identity *auth.Identity, resourcePath string, children bool) (*MetaData, error)

	// Remove removes a resource from the storage defined by resourcePath.
	Remove(identity *auth.Identity, resourcePath string, recursive bool) error

	// CreateContainer creates a container in the storage defined by resourcePath.
	CreateContainer(identity *auth.Identity, resourcePath string, recursive bool) error

	// Copy copies a resource from one resourcePath to another.
	// If resourcePaths belong to different storages this is a third party copy.
	Copy(identity *auth.Identity, fromPath, toPath string) error

	// Rename renames/move a resource from one resourcePath to another.
	// If resourcePaths belong to different storages this is a third party rename.
	Rename(identity *auth.Identity, fromPath, toPath string) error
}
