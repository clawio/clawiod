// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

// Capabilities reprents the capabilities of a storage
// Clients should ask for the capabilities of a storage before doing any operation.
type Capabilities struct {
	// Can upload a full object.
	PutObject bool `json:putobject`

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
	ThirdPartyRename bool `json:"thirdpartyrename"`

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

	// Can verify client checksums
	VerifyClientChecksum bool `json:"verifycientchecksum"`

	// Can send checksums to the client
	SendChecksum bool `json:"sendchecksum"`

	// The checksum supported on the server.
	SupportedChecksum string `json:"supportedchecksum"`

	// Create user home directory on login
	CreateUserHomeDirectory bool `json:"createuserhomedirectory"`
}
