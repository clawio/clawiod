// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

// Permissions represents the permissions over a resource.
// These permissions are based on Unix like permissions.
type Permissions struct {
	ReadCol     bool // Hability to read the name of the resources inside the collection
	WriteCol    bool // Hability to write, remove and rename resources inside the collection
	ExecuteCol  bool // Like read + metada about resources inside the collection
	ReadFile    bool // Hability to read a file
	WriteFile   bool // Hability to write to a file. Most of the storages implementations follow object store principle anddo not open files
	ExecuteFile bool // Hability to execute a file
}
