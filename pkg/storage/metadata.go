// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

// MetaData represents the metadata information about a resource.
type MetaData struct {
	ID           string       `json:"id"`           // The id of this resource.
	Path         string       `json:"path"`         // The path of this resource. It must be the full path with storage prefix like /home/cars/ford.png
	Size         uint64       `json:"size"`         // The size of this resource.
	IsContainer  bool         `json:"iscontainer"`  // Indicates if the resource is a container.
	MimeType     string       `json:"mimetype"`     // The mimetype of the resource.
	Checksum     string       `json:"checksum"`     // The checksum of the resource.
	ChecksumType string       `json:"checksumtype"` // The type of checksum used to calculate the checksum.
	Modified     uint64       `json:"modified"`     // The latest time the resource has been modified.
	ETag         string       `json:"etag"`         // The ETag http://en.wikipedia.org/wiki/HTTP_ETag.
	Permissions  *Permissions `json:"permissions"`  // The permissions for the resource.
	Children     []*MetaData  `json:"children"`     // If this resource is a container contains all the children´s metadata.
	Extra        interface{}  `json:"extra"`        // Contains extra attributes defined by the storage backend implementation.
}
