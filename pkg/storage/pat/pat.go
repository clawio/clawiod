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
	"io"
	"strings"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/storage"
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

	Capabilities(ctx context.Context, p *storage.CapabilitiesParams,
		pfx string) (*storage.Capabilities, error)
	CommitChunkedUpload(ctx context.Context, p *storage.CommitChunkUploadParams,
		rsp string) error
	Copy(ctx context.Context, p *storage.CopyParams) error
	CreateContainer(ctx context.Context, p *storage.CreateContainerParams) error
	CreateUserHomeDir(ctx context.Context, p *storage.CreateUserHomeDirParams, pfx string) error
	GetObject(ctx context.Context, p *storage.GetObjectParams) (io.Reader, error)
	PutChunkedObject(ctx context.Context, p *storage.PutChunkedObjectParams) error
	PutObject(ctx context.Context, p *storage.PutObjectParams) error
	Remove(ctx context.Context, p *storage.RemoveParams) error
	Rename(ctx context.Context, p *storage.RenameParams) error
	StartChunkedUpload(ctx context.Context, p *storage.StartChunkUploadParams,
		pfx string) (string, error)
	Stat(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error)
}

// pat implements the Pat interface.
type pat struct {
	storages map[string]storage.Storage
	cfg      config.Config
}

// NewParams are the params used by the New method.
type NewParams struct {
	Config config.Config
}

// New creates a Pat.
func New(p *NewParams) Pat {
	m := pat{
		storages: make(map[string]storage.Storage),
		cfg:      p.Config,
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

func (pt *pat) Capabilities(ctx context.Context, p *storage.CapabilitiesParams,
	pfx string) (*storage.Capabilities,
	error) {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return nil, err
	}
	return s.Capabilities(ctx, p), nil
}
func (pt *pat) CreateUserHomeDir(ctx context.Context, p *storage.CreateUserHomeDirParams,
	pfx string) error {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).CreateUserHomeDir {
		return &storage.NotImplementedError{
			Err: "CreateUserHomeDir not implemented",
		}
	}
	return s.CreateUserHomeDir(ctx, p)
}
func (pt *pat) PutObject(ctx context.Context, p *storage.PutObjectParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).PutObject {
		return &storage.NotImplementedError{
			Err: "PutObject not implemented",
		}
	}
	return s.PutObject(ctx, p)
}
func (pt *pat) StartChunkedUpload(ctx context.Context, p *storage.StartChunkUploadParams,
	pfx string) (string, error) {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return "", err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).PutObjectInChunks {
		return "", &storage.NotImplementedError{
			Err: "StartChunkedUpload not implemented",
		}
	}
	return s.StartChunkedUpload(ctx, p)
}
func (pt *pat) PutChunkedObject(ctx context.Context, p *storage.PutChunkedObjectParams) error {

	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).PutObjectInChunks {
		return &storage.NotImplementedError{
			Err: "PutChunkedObject not implemented",
		}
	}
	return s.PutChunkedObject(ctx, p)
}
func (pt *pat) CommitChunkedUpload(ctx context.Context, p *storage.CommitChunkUploadParams,
	pfx string) error {

	s, err := pt.getStorageFromPath(pfx)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).PutObjectInChunks {
		return &storage.NotImplementedError{
			Err: "PutObjectInChunks not implemented",
		}
	}
	return s.CommitChunkedUpload(ctx, p)
}
func (pt *pat) GetObject(ctx context.Context, p *storage.GetObjectParams) (io.Reader, error) {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return nil, err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).GetObject {
		return nil, &storage.NotImplementedError{
			Err: "GetObject not implemented",
		}
	}
	return s.GetObject(ctx, p)
}
func (pt *pat) Stat(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return nil, err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).Stat {
		return nil, &storage.NotImplementedError{
			Err: "Stat not implemented",
		}
	}
	return s.Stat(ctx, p)
}
func (pt *pat) Remove(ctx context.Context, p *storage.RemoveParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).Remove {
		return &storage.NotImplementedError{
			Err: "Remove not implemented",
		}
	}
	return s.Remove(ctx, p)
}
func (pt *pat) CreateContainer(ctx context.Context, p *storage.CreateContainerParams) error {
	s, err := pt.getStorageFromPath(p.Rsp)
	if err != nil {
		return err
	}
	cp := &storage.CapabilitiesParams{BaseParams: p.BaseParams}
	if !s.Capabilities(ctx, cp).CreateContainer {
		return &storage.NotImplementedError{
			Err: "CreateContainer not implemented",
		}
	}
	return s.CreateContainer(ctx, p)
}
func (pt *pat) Rename(ctx context.Context, p *storage.RenameParams) error {
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
	if !srcStrg.Capabilities(ctx, cp).Rename {
		return &storage.NotImplementedError{
			Err: "Rename not implemented",
		}
	}
	return srcStrg.Rename(ctx, p)
}

func (pt *pat) Copy(ctx context.Context, p *storage.CopyParams) error {
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
	if !srcStrg.Capabilities(ctx, cp).Copy {
		return &storage.NotImplementedError{
			Err: "Copy not implemented",
		}
	}
	return srcStrg.Copy(ctx, p)
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

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// patKey is the context key for the dispatcher.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const patKey key = 0

// NewContext returns a new Context carrying a storage pat.
func NewContext(ctx context.Context, p Pat) context.Context {
	return context.WithValue(ctx, patKey, p)
}

// FromContext extracts the storage pat from ctx, if present.
func FromContext(ctx context.Context) (Pat, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	p, ok := ctx.Value(patKey).(Pat)
	return p, ok
}

// MustFromContext extracts the storage pat from ctx.
// If not present it panics.
func MustFromContext(ctx context.Context) Pat {
	l, ok := ctx.Value(patKey).(Pat)
	if !ok {
		panic("storage pat is not registered")
	}
	return l
}
