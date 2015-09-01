// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package dispatcher defines the storage multiplexer to route storage operations against
// the registered storage providers.
package dispatcher

import (
	"fmt"
	"io"
	"net/url"

	"github.com/clawio/clawiod/lib/auth"
	"github.com/clawio/clawiod/lib/config"
	"github.com/clawio/clawiod/lib/logger"
	"github.com/clawio/clawiod/lib/storage"
)

// Dispatcher is the interface storage dispatchers must implement.
type Dispatcher interface {
	AddStorage(s storage.Storage) error
	GetStorage(storageScheme string) (storage.Storage, bool)
	GetAllStorages() []storage.Storage
	GetStoragesInfo() []*storage.Info
	IsUserHomeCreated(identity *auth.Identity, storageScheme string) (bool, error)
	CreateUserHome(identity *auth.Identity, storageScheme string) error
	PutFile(identity *auth.Identity, rawURI string, r io.Reader, size int64, checksumType, checksum string) error
	GetFile(identity *auth.Identity, rawURI string) (io.Reader, error)
	Stat(identity *auth.Identity, rawURI string, children bool) (*storage.MetaData, error)
	Remove(identity *auth.Identity, rawURI string, recursive bool) error
	CreateCol(identity *auth.Identity, rawURI string, recursive bool) error
	Copy(identity *auth.Identity, fromRawURI, toRawURI string) error
	Rename(identity *auth.Identity, fromRawURI, toRawURI string) error
}

// dispatcher dispatch storage operations to the correct storage
type dispatcher struct {
	storages map[string]storage.Storage
	cfg      *config.Config
	log      logger.Logger
}

// New creates a dispatcher or returns an error
func New(cfg *config.Config, log logger.Logger) Dispatcher {
	m := dispatcher{storages: make(map[string]storage.Storage), cfg: cfg, log: log}
	return &m
}

// AddStorage adds a storage
func (disp *dispatcher) AddStorage(s storage.Storage) error {
	if _, ok := disp.storages[s.GetScheme()]; ok {
		return fmt.Errorf("storage %s already registered", s.GetScheme())
	}
	disp.storages[s.GetScheme()] = s
	return nil
}

// GetStorage returns the storage with storageScheme and an boolean indicating if was found
func (disp *dispatcher) GetStorage(storageScheme string) (storage.Storage, bool) {
	sp, ok := disp.storages[storageScheme]
	return sp, ok
}

// GetAllStorages() []storage.Storage
func (disp *dispatcher) GetAllStorages() []storage.Storage {
	var storages []storage.Storage
	for _, s := range disp.storages {
		storages = append(storages, s)
	}
	return storages
}

// GetAllStorages() []storage.Storage
func (disp *dispatcher) GetStoragesInfo() []*storage.Info {
	var infos []*storage.Info
	storages := disp.GetAllStorages()
	for _, s := range storages {
		i := &storage.Info{
			Scheme:       s.GetScheme(),
			Capabilities: s.GetCapabilities(),
		}
		infos = append(infos, i)
	}
	return infos
}

// IsUserHomeCreated checks if the user home directory has been created in the specified storage.
func (disp *dispatcher) IsUserHomeCreated(identity *auth.Identity, storageScheme string) (bool, error) {
	strg, ok := disp.GetStorage(storageScheme)
	if !ok {
		return false, &storage.NotExistError{Err: fmt.Sprintf("storage '%s' not registered", storageScheme)}
	}
	ok, err := strg.IsUserHomeCreated(identity)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

// CreateUserHome routes the creation of the user home directory to the correct storage provider implementation.
// If the storageScheme is empty, the creation of the home directory will be propagated to all storages.
func (disp *dispatcher) CreateUserHome(identity *auth.Identity, storageScheme string) error {
	strg, ok := disp.GetStorage(storageScheme)
	if !ok {
		return &storage.NotExistError{Err: fmt.Sprintf("strg '%s' not registered", storageScheme)}
	}
	return strg.CreateUserHome(identity)
}

// PutFile routes the put operation to the correct storage provider implementation.
func (disp *dispatcher) PutFile(identity *auth.Identity, rawURI string, r io.Reader, size int64, checksumType, checksum string) error {
	s, uri, err := disp.getStorageAndURIFromPath(rawURI)
	if err != nil {
		return err
	}
	return s.PutFile(identity, uri, r, size, disp.cfg.GetDirectives().VerifyClientChecksum, checksum, checksumType)
}

// GetFile routes the get operation to the correct storage provider implementation.
func (disp *dispatcher) GetFile(identity *auth.Identity, rawURI string) (io.Reader, error) {
	s, uri, err := disp.getStorageAndURIFromPath(rawURI)
	if err != nil {
		return nil, err
	}
	return s.GetFile(identity, uri)
}

// Stat routes the stat operation to the correct storage provider implementation.
func (disp *dispatcher) Stat(identity *auth.Identity, rawURI string, children bool) (*storage.MetaData, error) {
	s, uri, err := disp.getStorageAndURIFromPath(rawURI)
	if err != nil {
		return nil, err
	}
	return s.Stat(identity, uri, children)
}

// Remove routes the remove operation to the correct storage provider implementation.
func (disp *dispatcher) Remove(identity *auth.Identity, rawURI string, recursive bool) error {
	s, uri, err := disp.getStorageAndURIFromPath(rawURI)
	if err != nil {
		return err
	}
	return s.Remove(identity, uri, recursive)
}

// CreateCol routes the create collection operation to the correct storage provider implementation.
func (disp *dispatcher) CreateCol(identity *auth.Identity, rawURI string, recursive bool) error {
	s, uri, err := disp.getStorageAndURIFromPath(rawURI)
	if err != nil {
		return err
	}
	return s.CreateCol(identity, uri, recursive)
}

func (disp *dispatcher) doCopyOrRename(op string, identity *auth.Identity, fromRawURI, toRawURI string) error {
	fromStorage, fromURI, err := disp.getStorageAndURIFromPath(fromRawURI)
	if err != nil {
		return err
	}
	toStorage, toURI, err := disp.getStorageAndURIFromPath(toRawURI)
	if err != nil {
		return err
	}
	if fromStorage.GetScheme() != toStorage.GetScheme() {
		return &storage.ThirdPartyCopyNotEnabled{}
	}
	if op == "copy" {
		return fromStorage.Copy(identity, fromURI, toURI)
	}
	return fromStorage.Rename(identity, fromURI, toURI)
}

// Copy routes the copy operation to the correct storage provider implementation.
func (disp *dispatcher) Copy(identity *auth.Identity, fromRawURI, toRawURI string) error {
	return disp.doCopyOrRename("copy", identity, fromRawURI, toRawURI)
}

// Rename routes the rename operation to the correct storage provider implementation.
func (disp *dispatcher) Rename(identity *auth.Identity, fromRawURI, toRawURI string) error {
	return disp.doCopyOrRename("rename", identity, fromRawURI, toRawURI)
}

// getStorageFromPath returns the storage provider adn the URI associated with the resourceURL passsed or an error.
// the resourceURL must be a well-formed URI like local://photos/beach.png or eos://data/big.dat
func (disp *dispatcher) getStorageAndURIFromPath(resourceURL string) (storage.Storage, *url.URL, error) {
	uri, err := url.Parse(resourceURL)
	if err != nil {
		return nil, nil, &storage.NotExistError{Err: err.Error()}
	}

	disp.log.Debugf("URI: %+v", map[string]interface{}{"url": resourceURL, "uri": fmt.Sprintf("%+v", *uri)})

	s, ok := disp.GetStorage(uri.Scheme)
	if !ok {
		return nil, nil, &storage.NotExistError{Err: fmt.Sprintf("storage '%s' not registered", uri.Scheme)}
	}
	uri.Opaque = ""
	return s, uri, nil
}
