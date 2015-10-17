// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package storage defines the Storage interface.
package storage

import (
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/clawio/clawiod/pkg/auth"
	"io"
	"strings"
)

const (
	// DefaultObjectMimeType is the default mime type for objects.
	DefaultObjectMimeType = "application/octet-stream"
	// DefaultContainerMimeType is the default mime type for containers.
	DefaultContainerMimeType = "application/container"
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
type Capabilities struct {
	// Can upload a full object.
	PutObject bool `json:"putobject"`

	// Can upload an object in chunks
	PutObjectInChunks bool `json:"putobjectinchunks"`

	// Can download a full object.
	GetObject bool `json:"getobject"`

	// Can download parts of an object
	GetObjectByByteRange bool `json:"getobjectbybyterange"`

	// Can get resource metadata.
	Stat bool `json:"stat"`

	// Can remove resources.
	Remove bool `json:"remove"`

	// Can create container.
	CreateContainer bool `json:"createcontainer"`

	// Can copy resources inside the same storage.
	Copy bool `json:"copy"`

	// Can rename resources inside the same storage.
	Rename bool `json:"rename"`

	// Can do third party copies.
	ThirdPartyCopy bool `json:"thirdpartycopy"`

	// Can do third party renames.
	ThirdPartyRename bool `json:"thridpartyrename"`

	// Can list the versions of a resource.
	ListVersions bool `json:"listversions"`

	// Can download a version.
	GetVersion bool `json:"getversion"`

	// Can create versions.
	CreateVersion bool `json:"createversion"`

	// Can rollback to a previous version.
	RollbackVersion bool `json:"rollbackversion"`

	// Can list deleted resources.
	ListDeletedResources bool `json:"listdeletedresources"`

	// Can restore a resource from the junk.
	RestoreDeletedResource bool `json:"restoredeletedresource"`

	// Can purge a resource.
	PurgeDeletedResource bool `json:"purgedeletedresource"`

	// Can send checksums to the client
	//SendChecksum bool

	// The checksum supported on the server.
	// If empty the server does not computes the checksum.
	// If this paremeter is set, the server will compute the checksum
	// independently of the value of VerifyClientChecksum.
	SupportedChecksum string `json:"supportedchecksum"`

	// Can verify client checksums.
	// When it is enabled and SupportedChecksum is defined,
	// the server will compare the client´s supplied checksum against
	// the server checksum.
	VerifyClientChecksum bool `json:"verifyclientchecksum"`

	// Create user home directory on login
	CreateUserHomeDir bool `json:"createuserhomedir"`
}

// Checksum represents the hexadecimal checksum.
type Checksum string

// Type returns the checksum type. Ex: md5
func (ck Checksum) Type() string {
	parts := strings.Split(ck.String(), ":")
	return parts[0]
}

// Value returns the value of the checksum. Ex: d71aa9d377eb509a82fe0511c4b7db50
func (ck Checksum) Value() string {
	parts := strings.Split(ck.String(), ":")
	if len(parts[0]) > 0 {
		return parts[1]
	}
	return ""
}

// String returns the string representation of the checksum.
func (ck Checksum) String() string {
	return string(ck)
}

// Range represents a byte-range.
type Range struct {
	Offset uint64
	Size   uint64
}

// MetaData represents the metadata information about a resource.
type MetaData struct {
	// The id of this resource.
	ID string `json:"id"`

	// The path of this resource.
	//It must be the full path with storage prefix like /home/cars/ford.png
	Path string `json:"path"`

	// The size of this resource.
	Size uint64 `json:"size"`

	// Indicates if the resource is a container.
	IsContainer bool `json:"iscontainer"`

	// The mimetype of the resource.
	MimeType string `json:"mimetype"`

	// The checksum of the resource.
	Checksum Checksum `json:"checksum"`

	// The latest time the resource has been modified.
	Modified uint64 `json:"modified"`

	// The ETag http://en.wikipedia.org/wiki/HTTP_ETag.
	ETag string `json:"etag"`

	// The permissions for the resource.
	Permissions ResourceMode `json:"permissions"`

	// If this resource is a container contains all the children´s metadata.
	Children []*MetaData `json:"children"`

	// Contains extra attributes defined by the storage backend implementation.
	// It can be useful to create custom applications.
	// An example could be the download redirection based on the user done
	// in CERNBox.
	Extra interface{} `json:"extra"`
}

// String returns the string representation of the metadata.
func (m *MetaData) String() string {
	return fmt.Sprintf("meta(%+v)", *m)
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

// BaseParams represents the base parameters for storage operations.
type BaseParams struct {
	// The Identity doing the operation.
	Idt *idm.Identity
	// Extra represents custom information to sent to the storage.
	Extra interface{}
}

// CapabilitiesParams are the params used by the Capabilities method.
type CapabilitiesParams struct {
	BaseParams
}

// CommitChunkUploadParams are the params used by ChunkUpload method.
type CommitChunkUploadParams struct {
	BaseParams
	Checksum Checksum
}

// CopyRenameParams are the common params used by Copy and Rename methods.
type CopyRenameParams struct {
	BaseParams
	Src string
	Dst string
}

// CopyParams are the params used by the Copy method.
type CopyParams struct {
	CopyRenameParams
}

// CreateContainerParams are the params used by the CreateContainer method.
type CreateContainerParams struct {
	BaseParams
	Rsp string
}

// CreateUserHomeDirParams are the params used by the CreateUserHomeDir method.
type CreateUserHomeDirParams struct {
	BaseParams
}

// GetObjectParams are the params used by the GetObject method.
type GetObjectParams struct {
	BaseParams
	Rsp   string
	Range *Range
	Size  uint64
}

// PutObjectCommonParams are the params used by the PutObject and
// PutChunkedObject methods.
type PutObjectCommonParams struct {
	BaseParams
	io.Reader
	Rsp  string
	Size uint64
}

// PutChunkedObjectParams are the params used by the PutChunkedObject method.
type PutChunkedObjectParams struct {
	PutObjectCommonParams
	TransferID string
}

// PutObjectParams are the params used by the PutObject method.
type PutObjectParams struct {
	PutObjectCommonParams
	Checksum Checksum
}

// RemoveParams are the params used by the Remove method.
type RemoveParams struct {
	BaseParams
	Rsp       string
	Recursive bool
}

// RenameParams are the params used by the Rename method.
type RenameParams struct {
	CopyRenameParams
}

// StartChunkUploadParams are the params used by the StartChunkUpload method.
type StartChunkUploadParams struct {
	BaseParams
}

// StatParams are the params used by the StatParams method.
type StatParams struct {
	BaseParams
	Rsp      string
	Children bool
}

// Storage is the interface that all the storage backends must implement
// to be used by the storage multiplexer.
type Storage interface {

	// GetCapabilities returns the capabilities of this storage.
	Capabilities(ctx context.Context, p *CapabilitiesParams) *Capabilities

	// CommitChunkedUpload commits the transaction
	CommitChunkedUpload(ctx context.Context, p *CommitChunkUploadParams) error

	// Copy copies a resource from one rsp to another.
	// If rsps belong to different storages this is a third party copy.
	Copy(ctx context.Context, p *CopyParams) error

	// CreateContainer creates a container in the storage
	// defined by rsp.
	CreateContainer(ctx context.Context, p *CreateContainerParams) error

	// CreateUserHomeDir creates the user home directory in the storage.
	CreateUserHomeDir(ctx context.Context, p *CreateUserHomeDirParams) error

	// GetObject gets an object from the storage defined by
	// the uri or by the resourceID.
	GetObject(ctx context.Context, p *GetObjectParams) (io.Reader, error)

	// Prefix returns the prefix of this storage.
	Prefix() string

	// PutChunkedObject uploads the chunk defined by start and
	// size of an object.
	PutChunkedObject(ctx context.Context, p *PutChunkedObjectParams) error

	// PutObject puts an object into the storage defined by rsp.
	PutObject(ctx context.Context, p *PutObjectParams) error

	// Remove removes a resource from the storage defined by rsp.
	Remove(ctx context.Context, p *RemoveParams) error

	// Rename renames/move a resource from one rsp to another.
	// If rsps belong to different storages this is
	// a third party rename.
	Rename(ctx context.Context, p *RenameParams) error

	// StartChunkedUpload starts a transaction for putting an object in chunks.
	StartChunkedUpload(ctx context.Context, p *StartChunkUploadParams) (string, error)

	// Stat returns metadata information about the resources and its children.
	Stat(ctx context.Context, p *StatParams) (*MetaData, error)
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

// Error returns the string representation of the error.
func (e *BadChecksumError) Error() string {
	msg := "data corrrupted. computed:%s and expected:%s"
	return fmt.Sprintf(msg, e.Computed, e.Expected)
}

// NotImplementedError represents the error of an un-implemented feature.
// Normally, the Storage Dispatcher will query the capabilities of a concrete
// storage before trigger the operation.
// The use of this error it should be due to custom logic in the storage.
type NotImplementedError struct {
	Err string
}

// Error returns the string representation of the error.
func (e *NotImplementedError) Error() string {
	return e.Err
}
