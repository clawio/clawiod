// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package root implements the root storage view.
package root

import (
	"io"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"

	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
)

type root struct {
	storagePrefix string
	storages      []storage.Storage
	config.Config
	logger.Logger
}

// NewParams are the params used by the New method.
type NewParams struct {
	Prefix   string
	Storages []storage.Storage
	Config   config.Config
}

// New creates a root object or returns an error.
func New(p *NewParams) storage.Storage {

	s := &root{storagePrefix: p.Prefix, storages: p.Storages,
		Config: p.Config}

	return s
}

func (s *root) Prefix() string {
	return s.storagePrefix
}

func (s *root) CreateUserHomeDir(ctx context.Context, p *storage.CreateUserHomeDirParams) error {
	return &storage.NotImplementedError{
		Err: "create user home directory not implemented in root storage"}
}

func (s *root) PutObject(ctx context.Context, p *storage.PutObjectParams) error {

	return &storage.NotImplementedError{
		Err: "put object not implemented in root storage"}
}
func (s *root) GetObject(ctx context.Context, p *storage.GetObjectParams) (io.Reader, error) {

	return nil, &storage.NotImplementedError{
		Err: "get object not implemented in root storage"}
}
func (s *root) Stat(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {

	ts := time.Now().UnixNano()
	tsString := strconv.FormatInt(ts, 10)

	parentMeta := &storage.MetaData{}
	parentMeta.ID = s.Prefix()
	parentMeta.Path = s.Prefix()
	parentMeta.IsContainer = true
	parentMeta.MimeType = storage.DefaultContainerMimeType
	parentMeta.Modified = uint64(ts)
	parentMeta.ETag = tsString

	if p.Children == false {
		return parentMeta, nil
	}

	parentMeta.Children = []*storage.MetaData{}
	for _, sto := range s.storages {
		m := &storage.MetaData{}
		m.ID = sto.Prefix()
		m.Path = m.ID
		m.IsContainer = true
		m.MimeType = storage.DefaultContainerMimeType
		m.Modified = uint64(ts)
		m.ETag = tsString
		parentMeta.Children = append(parentMeta.Children, m)
	}
	return parentMeta, nil
}

func (s *root) Remove(ctx context.Context, p *storage.RemoveParams) error {
	return &storage.NotImplementedError{
		Err: "remove not implemented in root storage"}
}

func (s *root) CreateContainer(ctx context.Context, p *storage.CreateContainerParams) error {
	return &storage.NotImplementedError{
		Err: "create container not implemented in root storage"}
}

func (s *root) Copy(ctx context.Context, p *storage.CopyParams) error {
	return &storage.NotImplementedError{
		Err: "copy not implemented in root storage"}
}

func (s *root) Rename(ctx context.Context, p *storage.RenameParams) error {
	return &storage.NotImplementedError{
		Err: "rename not implemented in root storage"}
}

func (s *root) StartChunkedUpload(ctx context.Context, p *storage.StartChunkUploadParams) (string, error) {
	return "", &storage.NotImplementedError{
		Err: "start chunk upload not implemented in root storage"}
}

func (s *root) PutChunkedObject(ctx context.Context, p *storage.PutChunkedObjectParams) error {
	return &storage.NotImplementedError{
		Err: "put chunked object not implemented in root storage"}
}

func (s *root) CommitChunkedUpload(ctx context.Context, p *storage.CommitChunkUploadParams) error {
	return &storage.NotImplementedError{
		Err: "commit chunked upload not implemented in root storage"}
}

func (s *root) Capabilities(ctx context.Context, p *storage.CapabilitiesParams) *storage.Capabilities {
	cap := &storage.Capabilities{}
	cap.Stat = true
	return cap
}
func (s *root) getPathWithoutStoragePrefix(rsp string) string {
	parts := strings.Split(rsp, "/")
	if len(parts) == 1 {
		return ""
	}
	return strings.Join(parts[1:], "/")
}
func (s *root) getPathWithStoragePrefix(relPath string) string {
	return path.Join(s.Prefix(), path.Clean(relPath))
}
