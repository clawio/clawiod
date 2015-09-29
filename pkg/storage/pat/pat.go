// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package pat defines the Pat interface and provides and implementation.
package pat

import (
	"fmt"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"io"
	"strings"
)

// Pat dispatchs storage operations.
type Pat interface {
	AddStorage(s storage.Storage) error
	GetStorage(resourcePath string) (storage.Storage, bool)
	GetAllStorages() []storage.Storage
	Capabilities(identity auth.Identity,
		prefix string) (storage.Capabilities, error)

	CommitChunkedUpload(chunkID string, checksum storage.Checksum) error
	Copy(identity auth.Identity, fromPath, toPath string) error
	CreateContainer(identity auth.Identity, resourcePath string) error
	CreateUserHomeDirectory(identity auth.Identity, resourcePath string) error
	GetObject(identity auth.Identity, resourcePath string) (io.Reader, error)
	PutChunkedObject(identity auth.Identity, r io.Reader, size int64,
		start int64, chunkID, resourcePath string) error

	PutObject(identity auth.Identity, resourcePath string, r io.Reader,
		size int64, checksum storage.Checksum) error

	Remove(identity auth.Identity, resourcePath string,
		recursive bool) error

	Rename(identity auth.Identity, fromPath, toPath string) error
	StartChunkedUpload(prefix string) (string, error)

	Stat(identity auth.Identity, resourcePath string,
		children bool) (storage.MetaData, error)
}

// New creates a Pat.
func New(cfg config.Config, log logger.Logger) Pat {
	m := pat{storages: make(map[string]storage.Storage), Config: cfg,
		Logger: log}

	return &m
}

func (p *pat) AddStorage(s storage.Storage) error {
	if _, ok := p.storages[s.Prefix()]; ok {
		return fmt.Errorf("pat: storage %s is already registered", s.Prefix())
	}
	p.storages[s.Prefix()] = s
	return nil
}

// GetStorage returns the storage with storageScheme and an boolean indicating if was found
func (p *pat) GetStorage(storageScheme string) (storage.Storage, bool) {
	sp, ok := p.storages[storageScheme]
	return sp, ok
}

// GetAllStorages returns all the storages registered.
func (p *pat) GetAllStorages() []storage.Storage {
	var storages []storage.Storage
	for _, s := range p.storages {
		storages = append(storages, s)
	}
	return storages
}

func (p *pat) Prefix() string {
	return ""
}
func (p *pat) Capabilities(identity auth.Identity,
	resourcePath string) (storage.Capabilities, error) {

	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return nil, err
	}
	return s.Capabilities(identity), nil
}
func (p *pat) CreateUserHomeDirectory(identity auth.Identity, resourcePath string) error {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.CreateUserHomeDirectory(identity)
}
func (p *pat) PutObject(identity auth.Identity, resourcePath string, r io.Reader, size int64, checksum storage.Checksum) error {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.PutObject(identity, resourcePath, r, size, checksum)
}
func (p *pat) StartChunkedUpload(prefix string) (string, error) {
	s, err := p.getStorageFromPath(prefix)
	if err != nil {
		return "", err
	}
	return s.StartChunkedUpload()
}
func (p *pat) PutChunkedObject(identity auth.Identity, r io.Reader, size int64,
	start int64, chunkID, resourcePath string) error {

	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.PutChunkedObject(identity, r, size, start, chunkID)
}
func (p *pat) CommitChunkedUpload(resourcePath string, checksum storage.Checksum) error {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.CommitChunkedUpload(checksum)
}
func (p *pat) GetObject(identity auth.Identity, resourcePath string) (io.Reader, error) {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return nil, err
	}
	return s.GetObject(identity, resourcePath)
}
func (p *pat) Stat(identity auth.Identity, resourcePath string, children bool) (storage.MetaData, error) {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return nil, err
	}
	return s.Stat(identity, resourcePath, children)
}
func (p *pat) Remove(identity auth.Identity, resourcePath string, recursive bool) error {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.Remove(identity, resourcePath, recursive)
}
func (p *pat) CreateContainer(identity auth.Identity, resourcePath string) error {
	s, err := p.getStorageFromPath(resourcePath)
	if err != nil {
		return err
	}
	return s.CreateContainer(identity, resourcePath)
}
func (p *pat) Rename(identity auth.Identity, fromPath, toPath string) error {
	fromStorage, err := p.getStorageFromPath(fromPath)
	if err != nil {
		return err
	}
	toStorage, err := p.getStorageFromPath(toPath)
	if err != nil {
		return err
	}
	if fromStorage.Prefix() != toStorage.Prefix() {
		return fmt.Errorf("third party rename from %s to %s not enabled yet", fromStorage.Prefix(), toStorage.Prefix())
	}
	return fromStorage.Rename(identity, fromPath, toPath)
}

func (p *pat) Copy(identity auth.Identity, fromPath, toPath string) error {
	fromStorage, err := p.getStorageFromPath(fromPath)
	if err != nil {
		return err
	}
	toStorage, err := p.getStorageFromPath(toPath)
	if err != nil {
		return err
	}
	if fromStorage.Prefix() != toStorage.Prefix() {
		return fmt.Errorf("third party copy from %s to %s not enabled yet", fromStorage.Prefix(), toStorage.Prefix())
	}
	return fromStorage.Rename(identity, fromPath, toPath)
}

// getStorageFromPath returns the storage implementation with the storage prefix used in resourcePath.
func (p *pat) getStorageFromPath(resourcePath string) (storage.Storage, error) {
	resourcePath = strings.TrimPrefix(resourcePath, "/")
	parts := strings.Split(resourcePath, "/")
	s, ok := p.GetStorage(parts[0])
	if !ok {
		return nil, &storage.NotExistError{Err: fmt.Sprintf("storage:%s not registered for resourcePath:%s", parts[0], resourcePath)}
	}
	return s, nil
}

// pat patchs storage operations to the correct storage
type pat struct {
	storages map[string]storage.Storage
	config.Config
	logger.Logger
}
