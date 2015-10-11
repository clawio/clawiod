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
	GetStorage(rsp string) (storage.Storage, bool)
	GetAllStorages() []storage.Storage
	Capabilities(idt auth.Identity,
		prfx string) (storage.Capabilities, error)

	CommitChunkedUpload(chunkID string, chk storage.Checksum) error
	Copy(idt auth.Identity, src, dst string) error
	CreateContainer(idt auth.Identity, rsp string) error
	CreateUserHomeDirectory(idt auth.Identity, rsp string) error
	GetObject(idt auth.Identity, rsp string, r *storage.Range) (io.Reader, error)
	PutChunkedObject(idt auth.Identity, r io.Reader, size int64,
		start int64, chunkID, rsp string) error

	PutObject(idt auth.Identity, rsp string, r io.Reader,
		size int64, chk storage.Checksum) error

	Remove(idt auth.Identity, rsp string,
		recursive bool) error

	Rename(idt auth.Identity, src, dst string) error
	StartChunkedUpload(prfx string) (string, error)

	Stat(idt auth.Identity, rsp string,
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
func (p *pat) Capabilities(idt auth.Identity,
	rsp string) (storage.Capabilities, error) {

	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return nil, err
	}
	return s.Capabilities(idt), nil
}
func (p *pat) CreateUserHomeDirectory(idt auth.Identity, rsp string) error {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.CreateUserHomeDirectory(idt)
}
func (p *pat) PutObject(idt auth.Identity, rsp string, r io.Reader, size int64, chk storage.Checksum) error {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.PutObject(idt, rsp, r, size, chk)
}
func (p *pat) StartChunkedUpload(prfx string) (string, error) {
	s, err := p.getStorageFromPath(prfx)
	if err != nil {
		return "", err
	}
	return s.StartChunkedUpload()
}
func (p *pat) PutChunkedObject(idt auth.Identity, r io.Reader, size int64,
	start int64, chunkID, rsp string) error {

	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.PutChunkedObject(idt, r, size, start, chunkID)
}
func (p *pat) CommitChunkedUpload(rsp string, chk storage.Checksum) error {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.CommitChunkedUpload(chk)
}
func (p *pat) GetObject(idt auth.Identity, rsp string, r *storage.Range) (io.Reader, error) {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return nil, err
	}
	return s.GetObject(idt, rsp, r)
}
func (p *pat) Stat(idt auth.Identity, rsp string, children bool) (storage.MetaData, error) {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return nil, err
	}
	return s.Stat(idt, rsp, children)
}
func (p *pat) Remove(idt auth.Identity, rsp string, recursive bool) error {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.Remove(idt, rsp, recursive)
}
func (p *pat) CreateContainer(idt auth.Identity, rsp string) error {
	s, err := p.getStorageFromPath(rsp)
	if err != nil {
		return err
	}
	return s.CreateContainer(idt, rsp)
}
func (p *pat) Rename(idt auth.Identity, src, dst string) error {
	srcStrg, err := p.getStorageFromPath(src)
	if err != nil {
		return err
	}
	dstStrg, err := p.getStorageFromPath(dst)
	if err != nil {
		return err
	}
	if srcStrg.Prefix() != dstStrg.Prefix() {
		return fmt.Errorf("third party rename from %s to %s not enabled yet", srcStrg.Prefix(), dstStrg.Prefix())
	}
	return srcStrg.Rename(idt, src, dst)
}

func (p *pat) Copy(idt auth.Identity, src, dst string) error {
	srcStrg, err := p.getStorageFromPath(src)
	if err != nil {
		return err
	}
	dstStrg, err := p.getStorageFromPath(dst)
	if err != nil {
		return err
	}
	if srcStrg.Prefix() != dstStrg.Prefix() {
		return fmt.Errorf("third party copy from %s to %s not enabled yet", srcStrg.Prefix(), dstStrg.Prefix())
	}
	return srcStrg.Rename(idt, src, dst)
}

// getStorageFromPath returns the storage implementation with the storage prfx used in rsp.
func (p *pat) getStorageFromPath(rsp string) (storage.Storage, error) {
	rsp = strings.TrimPrefix(rsp, "/")
	parts := strings.Split(rsp, "/")
	s, ok := p.GetStorage(parts[0])
	if !ok {
		return nil, &storage.NotExistError{Err: fmt.Sprintf("storage:%s not registered for rsp:%s", parts[0], rsp)}
	}
	return s, nil
}

// pat patchs storage operations to the correct storage
type pat struct {
	storages map[string]storage.Storage
	config.Config
	logger.Logger
}
