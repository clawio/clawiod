// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package dispatcher defines the storage dispatcher to route data and metadata operations
// to the corresponding storage implementation.
package dispatcher

import (
	"fmt"
	"github.com/clawio/clawiod/pkg/logger"
	"io"
	"strings"

	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/storage"
)

// Dispatcher is the interface storage dispatchers must implement.
type Dispatcher interface {
	AddStorage(s storage.Storage) error
	GetStorage(path string) (storage.Storage, bool)
	GetAllStorages() []storage.Storage

	DispatchGetCapabilities(identity *auth.Identity, path string) (*storage.Capabilities, error)
	DispatchCreateUserHomeDirectory(identity *auth.Identity, path string) error
	DispatchPutObject(identity *auth.Identity, path string, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error
	DispatchStartChunkedUpload(path string) (string, error)
	DispatchPutChunkedObject(identity *auth.Identity, r io.Reader, size int64, start int64, storagePrefix, chunkID string) error
	DispatchCommitChunkedUpload(path, chunkID string, verifyChecksum bool, checksum, checksumType string) error
	DispatchGetObject(identity *auth.Identity, path string) (io.Reader, error)
	DispatchStat(identity *auth.Identity, path string, children bool) (*storage.MetaData, error)
	DispatchRemove(identity *auth.Identity, path string, recursive bool) error
	DispatchCreateContainer(identity *auth.Identity, path string, recursive bool) error
	DispatchRename(identity *auth.Identity, fromPath, toPath string) error
	DispatchCopy(identity *auth.Identity, fromPath, toPath string) error
}

// dispatcher dispatchs storage operations to the correct storage
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
	if _, ok := disp.storages[s.GetStoragePrefix()]; ok {
		return fmt.Errorf("storage:%s is already registered", s.GetStoragePrefix())
	}
	disp.storages[s.GetStoragePrefix()] = s
	return nil
}

// GetStorage returns the storage with storageScheme and an boolean indicating if was found
func (disp *dispatcher) GetStorage(storageScheme string) (storage.Storage, bool) {
	sp, ok := disp.storages[storageScheme]
	return sp, ok
}

// GetAllStorages returns all the storages registered.
func (disp *dispatcher) GetAllStorages() []storage.Storage {
	var storages []storage.Storage
	for _, s := range disp.storages {
		storages = append(storages, s)
	}
	return storages
}

func (disp *dispatcher) DispatchGetCapabilities(identity *auth.Identity, path string) (*storage.Capabilities, error) {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return nil, err
	}
	return s.GetCapabilities(identity), nil
}
func (disp *dispatcher) DispatchCreateUserHomeDirectory(identity *auth.Identity, path string) error {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return err
	}
	return s.CreateUserHomeDirectory(identity)
}
func (disp *dispatcher) DispatchPutObject(identity *auth.Identity, path string, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return err
	}
	return s.PutObject(identity, path, r, size, verifyChecksum, checksum, checksumType)
}
func (disp *dispatcher) DispatchStartChunkedUpload(storagePrefix string) (string, error) {
	s, err := disp.getStorageFromPath(storagePrefix)
	if err != nil {
		return "", err
	}
	return s.StartChunkedUpload()
}
func (disp *dispatcher) DispatchPutChunkedObject(identity *auth.Identity, r io.Reader, size int64, start int64, storagePrefix, chunkID string) error {
	s, err := disp.getStorageFromPath(storagePrefix)
	if err != nil {
		return err
	}
	return s.PutChunkedObject(identity, r, size, start, chunkID)
}
func (disp *dispatcher) DispatchCommitChunkedUpload(path, chunkID string, verifyChecksum bool, checksum, checksumType string) error {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return err
	}
	return s.CommitChunkedUpload(chunkID, verifyChecksum, checksum, checksumType)
}
func (disp *dispatcher) DispatchGetObject(identity *auth.Identity, path string) (io.Reader, error) {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return nil, err
	}
	return s.GetObject(identity, path)
}
func (disp *dispatcher) DispatchStat(identity *auth.Identity, path string, children bool) (*storage.MetaData, error) {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return nil, err
	}
	return s.Stat(identity, path, children)
}
func (disp *dispatcher) DispatchRemove(identity *auth.Identity, path string, recursive bool) error {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return err
	}
	return s.Remove(identity, path, recursive)
}
func (disp *dispatcher) DispatchCreateContainer(identity *auth.Identity, path string, recursive bool) error {
	s, err := disp.getStorageFromPath(path)
	if err != nil {
		return err
	}
	return s.CreateContainer(identity, path, recursive)
}
func (disp *dispatcher) DispatchRename(identity *auth.Identity, fromPath, toPath string) error {
	fromStorage, err := disp.getStorageFromPath(fromPath)
	if err != nil {
		return err
	}
	toStorage, err := disp.getStorageFromPath(toPath)
	if err != nil {
		return err
	}
	if fromStorage.GetStoragePrefix() != toStorage.GetStoragePrefix() {
		return fmt.Errorf("third party rename from %s to %s not enabled yet", fromStorage.GetStoragePrefix(), toStorage.GetStoragePrefix())
	}
	return fromStorage.Rename(identity, fromPath, toPath)
}

func (disp *dispatcher) DispatchCopy(identity *auth.Identity, fromPath, toPath string) error {
	fromStorage, err := disp.getStorageFromPath(fromPath)
	if err != nil {
		return err
	}
	toStorage, err := disp.getStorageFromPath(toPath)
	if err != nil {
		return err
	}
	if fromStorage.GetStoragePrefix() != toStorage.GetStoragePrefix() {
		return fmt.Errorf("third party copy from %s to %s not enabled yet", fromStorage.GetStoragePrefix(), toStorage.GetStoragePrefix())
	}
	return fromStorage.Rename(identity, fromPath, toPath)
}

// getStorageFromPath returns the storage implementation with the storage prefix used in path.
func (disp *dispatcher) getStorageFromPath(path string) (storage.Storage, error) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	s, ok := disp.GetStorage(parts[0])
	if !ok {
		return nil, &storage.NotExistError{Err: fmt.Sprintf("storage:%s not registered for path:%s", parts[0], path)}
	}
	return s, nil
}
