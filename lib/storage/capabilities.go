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
type Capabilities struct {
	// Allowed to upload files.
	PutFile bool `json:put`

	// Allowed to upload files in chunks
	PutFileChunked bool `json:"put_chunked"`

	// Allowed to download resources.
	GetFile bool `json:"get"`

	// Allowed to Seek (for Byte Range Requests)
	GetFileByteRange bool `json:"get_byte_range"`

	// Allowed to get metadata information.
	Stat bool `json:"stat"`

	// Allowed to remove resources.
	Remove bool `json:"remove"`

	// Allowed to create collections.
	Mkcol bool `json:"mkcol"`

	// Allowed to copy resources inside the same storage.
	Copy bool `json:"copy"`

	// Allowed to rename resources inside the same storage.
	Rename bool `json:"rename"`

	// Allowed to do third party copies (src and dst storage have different schemes).
	ThirdPartyCopy bool `json:"3rd_party_copy"`

	// Allowed to do third party renames (src and dst storage have different schemes).
	ThirdPartyRename bool `json:"3rd_party_rename"`

	// Allowed to list the versions of a resource.
	ListVersions bool `json:"list_versions"`

	// Allowed to download a version.
	GetVersion bool `json:"get_version"`

	// Allow the user to trigger the creation of a version.
	// The storage is reponsible to create new versions but the user can have the permission
	// to trigger on-demand version.
	CreateVersion bool `json:"create_version"`

	// Allowed to rollback to a previous version.
	RollbackVersion bool `json:"rollback_version"`

	// Allowed to list the resources in the junk.
	ListJunkFiles bool `json:"list_junk_files"`

	// Allowed to download a resource in the junk.
	GetJunkFile bool `json:"get_junk_file"`

	// Allowed to restore a resource from the junk.
	RestoreJunkFile bool `json:"restore_junk_file"`

	// Allowed to purge a resource from the junk. Remove completely.
	PurgeJunkFile bool `purge_junk_file`
}
