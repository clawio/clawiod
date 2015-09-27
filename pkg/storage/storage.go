// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package storage defines the interface that storage providers
// must implement and
// defines the metadata resource.
package storage

import (
	"fmt"
	"github.com/clawio/clawiod/pkg/auth"
	"io"
	"strings"
)

// The first 16 bits are reserved and are part of the public API.
// The last 16 bits are available for extending resource permissions and
// are not part of the public API.
//     +-----+-------+-------------------------------------------------+--+
//     | Bit | Name  | Meaning                                         |  |
//     +-----+-------+-------------------------------------------------+--+
//     | 0   | PADD  | Grants adding resources to a container.         |  |
//     | 1   | PFSHR | Grants federated sharing for the container.     |  |
//     | 2   | PGET  | Grants downloading an object.                   |  |
//     | 3   | PLINK | Grants creating links from a resource.          |  |
//     | 4   | PLIST | Grants listing of resources inside a container. |  |
//     | 5   | PRM   | Grants removing a resource.                     |  |
//     | 6   | PSHR  | Grants internal sharing for the container.      |  |
//     | 7   | PSTAT | Grants getting metadata from a resource.        |  |
//     | 8   | -     | RESERVED                                        |  |
//     | 9   | -     | RESERVED                                        |  |
//     | 10  | -     | RESERVED                                        |  |
//     | 11  | -     | RESERVED                                        |  |
//     | 12  | -     | RESERVED                                        |  |
//     | 13  | -     | RESERVED                                        |  |
//     | 14  | -     | RESERVED                                        |  |
//     | 15  | -     | RESERVED                                        |  |
//     | 16  | -     | AVAILABLE                                       |  |
//     | 17  | -     | AVAILABLE                                       |  |
//     | 18  | -     | AVAILABLE                                       |  |
//     | 19  | -     | AVAILABLE                                       |  |
//     | 20  | -     | AVAILABLE                                       |  |
//     | 21  | -     | AVAILABLE                                       |  |
//     | 22  | -     | AVAILABLE                                       |  |
//     | 23  | -     | AVAILABLE                                       |  |
//     | 24  | -     | AVAILABLE                                       |  |
//     | 25  | -     | AVAILABLE                                       |  |
//     | 26  | -     | AVAILABLE                                       |  |
//     | 27  | -     | AVAILABLE                                       |  |
//     | 28  | -     | AVAILABLE                                       |  |
//     | 29  | -     | AVAILABLE                                       |  |
//     | 30  | -     | AVAILABLE                                       |  |
//     | 31  | -     | AVAILABLE                                       |  |
//     +-----+-------+-------------------------------------------------+--+
const (
	PADD ResourceMode = 1 << (32 - 1 - iota)
	PFSHR
	PGET
	PLINK
	PLIST
	PRM
	PSHR
	PSTAT
)

// Capabilities represents the capabilities of a storage.
// Clients should ask for the capabilities of a storage
// before doing any operation.
type Capabilities interface {
	// Can upload a full object.
	PutObject() bool

	// Can upload an object in chunks
	PutObjectInChunks() bool

	// Can download a full object.
	GetObject() bool

	// Can download parts of an object
	GetObjectByByteRange() bool

	// Can get resource metadata.
	Stat() bool

	// Can remove resources.
	Remove() bool

	// Can create container.
	CreateContainer() bool

	// Can copy resources inside the same storage.
	Copy() bool

	// Can rename resources inside the same storage.
	Rename() bool

	// Can do third party copies.
	ThirdPartyCopy() bool

	// Can do third party renames.
	ThirdPartyRename() bool

	// Can list the versions of a resource.
	ListVersions() bool

	// Can download a version.
	GetVersion() bool

	// Can create versions.
	CreateVersion() bool

	// Can rollback to a previous version.
	RollbackVersion() bool

	// Can list deleted resources.
	ListDeletedResources() bool

	// Can restore a resource from the junk.
	RestoreDeletedResource() bool

	// Can purge a resource.
	PurgeDeletedResource() bool

	// Can verify client checksums
	VerifyClientChecksum() bool

	// Can send checksums to the client
	SendChecksum() bool

	// The checksum supported on the server.
	SupportedChecksum() string

	// Create user home directory on login
	CreateUserHomeDirectory() bool
}

type Checksum string

func (ck Checksum) Type() string {
	parts := strings.Split(ck.String(), ":")
	return parts[0]
}

func (ck Checksum) Value() string {
	parts := strings.Split(ck.String(), ":")
	if len(parts[0]) > 0 {
		return parts[1]
	}
	return ""
}
func (ck Checksum) String() string {
	return string(ck)
}

// MetaData represents the metadata information about a resource.
type MetaData interface {
	// The id of this resource.
	ID() string

	// The path of this resource.
	//It must be the full path with storage prefix like /home/cars/ford.png
	Path() string

	// The size of this resource.
	Size() uint64

	// Indicates if the resource is a container.
	IsContainer() bool

	// The mimetype of the resource.
	MimeType() string

	// The checksum of the resource.
	Checksum() Checksum

	// The latest time the resource has been modified.
	Modified() uint64

	// The ETag http://en.wikipedia.org/wiki/HTTP_ETag.
	ETag() string

	// The permissions for the resource.
	Permissions() ResourceMode

	// If this resource is a container contains all the children´s metadata.
	Children() []MetaData

	// Contains extra attributes defined by the storage backend implementation.
	// It can be useful to create custom applications.
	// An example could be the download redirection based on the user done
	// in CERNBox.
	Extra() interface{}
}

// A ResourceMode represents a resource's permission bits.
type ResourceMode uint32

// IsAddable reports whether a container can add objects.
// That is, it tests that PADD bit is set.
func (m ResourceMode) IsAddable() bool {
	return m&PADD != 0
}

// IsFederatedShareable reports whether a container can be shared
// with a federated entity.
// That is, it tests that PFSHR bit is set.
func (m ResourceMode) IsFederatedShareable() bool {
	return m&PFSHR != 0
}

// IsGettable reports whether an object can be downloaded.
// That is, it tests that PGET bit is set.
func (m ResourceMode) IsGettable() bool {
	return m&PGET != 0
}

// IsLinkable reports whether a resource can be linked.
// That is, it tests that PLINK bit is set.
func (m ResourceMode) IsLinkable() bool {
	return m&PLINK != 0
}

// IsListable reports whether a container can be listed.
// That is, it tests that PLIST bit is set.
func (m ResourceMode) IsListable() bool {
	return m&PLIST != 0
}

// IsRemovable reports whether a resource can be removed.
// That is, it tests that PRM bit is set.
func (m ResourceMode) IsRemovable() bool {
	return m&PRM != 0
}

// IsShareable reports whether a container can be shared.
// That is, it tests that PSHR bit is set.
func (m ResourceMode) IsShareable() bool {
	return m&PSHR != 0
}

// IsStatable reports whether a a resource metadata can be obtained.
// That is, it tests that PSHR bit is set.
func (m ResourceMode) IsStatable() bool {
	return m&PSTAT != 0
}

// Storage is the interface that all the storage backends must implement
// to be used by the storage multiplexer.
type Storage interface {

	// GetCapabilities returns the capabilities of this storage.
	Capabilities(identity auth.Identity) Capabilities

	// CommitChunkedUpload commits the transaction
	CommitChunkedUpload(chunkID string, checksum, checksumType string) error

	// Copy copies a resource from one resourcePath to another.
	// If resourcePaths belong to different storages this is a third party copy.
	Copy(identity auth.Identity, fromPath, toPath string) error

	// CreateContainer creates a container in the storage
	// defined by resourcePath.
	CreateContainer(identity auth.Identity, resourcePath string) error

	// CreateUserHomeDirectory creates the user home directory in the storage.
	CreateUserHomeDirectory(identity auth.Identity) error

	// GetObject gets an object from the storage defined by
	// the uri or by the resourceID.
	GetObject(identity auth.Identity, resourcePath string) (io.Reader, error)

	// Prefix returns the prefix of this storage.
	Prefix() string

	// PutChunkedObject uploads the chunk defined by start and
	// size of an object.
	PutChunkedObject(identity auth.Identity, r io.Reader, size int64,
		start int64, chunkID string) error

	// PutObject puts an object into the storage defined by resourcePath.
	PutObject(identity auth.Identity, resourcePath string, r io.Reader,
		size int64, checksum Checksum) error

	// Remove removes a resource from the storage defined by resourcePath.
	Remove(identity auth.Identity, resourcePath string,
		recursive bool) error

	// Rename renames/move a resource from one resourcePath to another.
	// If resourcePaths belong to different storages this is
	//  a third party rename.
	Rename(identity auth.Identity, fromPath, toPath string) error

	// StartChunkedUpload starts a transaction for putting an object in chunks.
	StartChunkedUpload() (string, error)

	// Stat returns metadata information about the resources and its children.
	Stat(identity auth.Identity, resourcePath string,
		children bool) (MetaData, error)
}

// AlreadyExistError represents the error
// ocurred when the resource already exists.
type AlreadyExistError struct {
	Err string
}

func (e *AlreadyExistError) Error() string { return e.Err }

// NotExistError represents the error ocurred
// when the resource does not exist.
type NotExistError struct {
	Err string
}

func (e *NotExistError) Error() string { return e.Err }

// BadChecksumError represents the error of a missmatch
// between client and server checksums.
type BadChecksumError struct {
	Expected string
	Computed string
}

func (e *BadChecksumError) Error() string {
	msg := "data corrrupted. computed:%s and expected:%s"
	return fmt.Sprintf(msg, e.Computed, e.Expected)
}

type NotImplementedError struct {
	Err string
}

func (e *NotImplementedError) Error() string {
	return e.Err
}
