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
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"io"
	"strings"
)

// BaseParams are the params used in Pat operations.
type BaseParams struct {
	LID   string
	Extra interface{}
}

// AddStorageParams are the params used by the AddStorage method.
type AddStorageParams struct {
	BaseParams
	Storage storage.Storage
}

// GetStorageParams are the params used by the GetStorage method.
type GetStorageParams struct {
	BaseParams
	Rsp string
}

// GetAllStoragesParams are the params used by the GetAllStorages method.
type GetAllStoragesParams struct {
	BaseParams
}

// Pat dispatchs storage operations to concrete storage implementations based
// on the storage prefix.
type Pat interface {
	AddStorage(p *AddStorageParams) error
	GetStorage(p *GetStorageParams) (storage.Storage, bool)
	GetAllStorages(p *GetAllStoragesParams) []storage.Storage

	Capabilities(p *storage.CapabilitiesParams,
		pfx string) (*storage.Capabilities, error)
	CommitChunkedUpload(p *storage.CommitChunkUploadParams,
		rsp string) error
	Copy(p *storage.CopyParams) error
	CreateContainer(p *storage.CreateContainerParams) error
	CreateUserHomeDir(p *storage.CreateUserHomeDirParams, pfx string) error
	GetObject(ip *storage.GetObjectParams) (io.Reader, error)
	PutChunkedObject(p *storage.PutChunkedObjectParams) error
	PutObject(p *storage.PutObjectParams) error
	Remove(p *storage.RemoveParams) error
	Rename(p *storage.RenameParams) error
	StartChunkedUpload(p *storage.StartChunkUploadParams,
		pfx string) (string, error)
	Stat(p *storage.StatParams) (*storage.MetaData, error)
}

// pat implements the Pat interface.
type pat struct {
	storages map[string]storage.Storage
	log      logger.Logger
	cfg      config.Config
}

// NewParams are the params used by the New method.
type NewParams struct {
	Config config.Config
	Log    logger.Logger
}

// New creates a Pat.
func New(p *NewParams) Pat {
	m := pat{
		storages: make(map[string]storage.Storage),
		cfg:      p.Config,
		log:      p.Log,
	}
	return &m
}

func (pt *pat) AddStorage(p *AddStorageParams) error {
	if _, ok := pt.storages[p.Storage.Prefix()]; ok {
		return fmt.Errorf("AddStorage: storage %s is already registered",
			p.Storage.Prefix())
	}
	pt.storages[p.Storage.Prefix()] = p.Storage
	return nil
}

// GetStorage returns the storage with storageScheme
// and an boolean indicating if was found
func (pt *pat) GetStorage(p *GetStorageParams) (storage.Storage, bool) {
	s, ok := pt.storages[p.Rsp]
	return s, ok
}

// GetAllStorages returns all the storages registered.
func (pt *pat) GetAllStorages(p *GetAllStoragesParams) []storage.Storage {
	var storages []storage.Storage
	for _, s := range pt.storages {
		storages = append(storages, s)
	}
	return storages
}

func (pt *pat) Capabilities(p *storage.CapabilitiesParams,
	pfx string) (*storage.Capabilities,
	error) {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return nil, err
	}
	return s.Capabilities(p), nil
}
func (pt *pat) CreateUserHomeDir(p *storage.CreateUserHomeDirParams,
	pfx string) error {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).CreateUserHomeDir {
		return &storage.NotImplementedError{
			Err: "CreateUserHomeDir not implemented",
		}
	}
	return s.CreateUserHomeDir(p)
}
func (pt *pat) PutObject(p *storage.PutObjectParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).PutObject {
		return &storage.NotImplementedError{
			Err: "PutObject not implemented",
		}
	}
	return s.PutObject(p)
}
func (pt *pat) StartChunkedUpload(p *storage.StartChunkUploadParams,
	pfx string) (string, error) {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return "", err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).PutObjectInChunks {
		return "", &storage.NotImplementedError{
			Err: "StartChunkedUpload not implemented",
		}
	}
	return s.StartChunkedUpload(p)
}
func (pt *pat) PutChunkedObject(p *storage.PutChunkedObjectParams) error {

	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).PutObjectInChunks {
		return &storage.NotImplementedError{
			Err: "PutChunkedObject not implemented",
		}
	}
	return s.PutChunkedObject(p)
}
func (pt *pat) CommitChunkedUpload(p *storage.CommitChunkUploadParams,
	pfx string) error {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).PutObjectInChunks {
		return &storage.NotImplementedError{
			Err: "PutObjectInChunks not implemented",
		}
	}
	return s.CommitChunkedUpload(p)
}
func (pt *pat) GetObject(p *storage.GetObjectParams) (io.Reader, error) {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return nil, err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).GetObject {
		return nil, &storage.NotImplementedError{
			Err: "GetObject not implemented",
		}
	}
	return s.GetObject(p)
}
func (pt *pat) Stat(p *storage.StatParams) (*storage.MetaData, error) {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return nil, err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).Stat {
		return nil, &storage.NotImplementedError{
			Err: "Stat not implemented",
		}
	}
	return s.Stat(p)
}
func (pt *pat) Remove(p *storage.RemoveParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).Remove {
		return &storage.NotImplementedError{
			Err: "Remove not implemented",
		}
	}
	return s.Remove(p)
}
func (pt *pat) CreateContainer(p *storage.CreateContainerParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(cp).CreateContainer {
		return &storage.NotImplementedError{
			Err: "CreateContainer not implemented",
		}
	}
	return s.CreateContainer(p)
}
func (pt *pat) Rename(p *storage.RenameParams) error {
	srcStrg, err := pt.getStorageFromPath(p.Src)
	if err != nil {
		return err
	}
	dstStrg, err := pt.getStorageFromPath(p.Dst)
	if err != nil {
		return err
	}
	if srcStrg.Prefix() != dstStrg.Prefix() {
		return fmt.Errorf("Third party rename from %s to %s not enabled yet",
			srcStrg.Prefix(), dstStrg.Prefix())
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !srcStrg.Capabilities(cp).Rename {
		return &storage.NotImplementedError{
			Err: "Rename not implemented",
		}
	}
	return srcStrg.Rename(p)
}

func (pt *pat) Copy(p *storage.CopyParams) error {
	srcStrg, err := pt.getStorageFromPath(p.Src)
	if err != nil {
		return err
	}
	dstStrg, err := pt.getStorageFromPath(p.Dst)
	if err != nil {
		return err
	}
	if srcStrg.Prefix() != dstStrg.Prefix() {
		return fmt.Errorf("third party copy from %s to %s not enabled yet",
			srcStrg.Prefix(), dstStrg.Prefix())
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !srcStrg.Capabilities(cp).Copy {
		return &storage.NotImplementedError{
			Err: "Copy not implemented",
		}
	}
	return srcStrg.Copy(p)
}

// getStorageFromPath returns the storage implementation with the storage prfx used in rspt.
func (pt *pat) getStorageFromPath(rsp string) (storage.Storage, error) {
	rsp = strings.TrimPrefix(rsp, "/")
	parts := strings.Split(rsp, "/")
	p := &GetStorageParams{}
	p.Rsp = parts[0]
	s, ok := pt.GetStorage(p)
	if !ok {
		return nil, &storage.NotExistError{Err: fmt.Sprintf("storage:%s not registered for rsp:%s", parts[0], rsp)}
	}
	return s, nil
}
