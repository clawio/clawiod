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
type Permissions struct {
	Stat           bool `json:"stat"`           // Grants getting metadata from a resource.
	List           bool `json:"list"`           // Grants listing of resources inside a container.
	Add            bool `json:"add"`            // Grants adding resources to the container.
	Get            bool `json:"get"`            // Grants downloading an object.
	Remove         bool `json:"remove"`         // Grants removing a resource.
	Link           bool `json:"link"`           // Grants creating links from a resource.
	Share          bool `json:"share"`          // Grants internal sharing for the container.
	FederatedShare bool `json:"federatedshare"` // Grants federated sharing for the container
}
