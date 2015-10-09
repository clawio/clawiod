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

	"github.com/clawio/clawiod/pkg/auth"
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

// New creates a root object or returns an error.
func New(storagePrefix string, storages []storage.Storage,
	cfg config.Config, log logger.Logger) storage.Storage {

	s := &root{storagePrefix: storagePrefix, storages: storages,
		Config: cfg, Logger: log}

	return s
}

func (s *root) Prefix() string {
	return s.storagePrefix
}

func (s *root) CreateUserHomeDirectory(identity auth.Identity) error {
	return &storage.NotImplementedError{
		"create user home directory not implemented in root storage",
	}
}

func (s *root) PutObject(identity auth.Identity, resourcePath string,
	r io.Reader, size int64, checksum storage.Checksum) error {

	return &storage.NotImplementedError{
		"put object not implemented in root storage",
	}
}
func (s *root) GetObject(identity auth.Identity,
	resourcePath string) (io.Reader, error) {

	return nil, &storage.NotImplementedError{
		"get object not implemented in root storage",
	}
}
func (s *root) Stat(identity auth.Identity, resourcePath string,
	children bool) (storage.MetaData, error) {

	ts := time.Now().UnixNano()
	tsString := strconv.FormatInt(ts, 10)

	parentMeta := &meta{}
	parentMeta.id = s.Prefix()
	parentMeta.path = s.Prefix()
	parentMeta.isContainer = true
	parentMeta.mimeType = storage.DEFAULT_CONTAINER_MIMETYPE
	parentMeta.modified = uint64(ts)
	parentMeta.etag = tsString

	if children == false {
		return parentMeta, nil
	}

	parentMeta.children = []storage.MetaData{}
	for _, sto := range s.storages {
		m := &meta{}
		m.id = sto.Prefix()
		m.path = m.id
		m.isContainer = true
		m.mimeType = storage.DEFAULT_CONTAINER_MIMETYPE
		m.modified = uint64(ts)
		m.etag = tsString
		parentMeta.children = append(parentMeta.children, m)
	}

	return parentMeta, nil
}

func (s *root) Remove(identity auth.Identity, resourcePath string, recursive bool) error {
	return &storage.NotImplementedError{"remove not implemented in root storage"}
}

func (s *root) CreateContainer(identity auth.Identity, resourcePath string) error {
	return &storage.NotImplementedError{"create container not implemented in root storage"}
}

func (s *root) Copy(identity auth.Identity, fromPath, toPath string) error {
	return &storage.NotImplementedError{"copy not implemented in root storage"}
}

func (s *root) Rename(identity auth.Identity, fromPath, toPath string) error {
	return &storage.NotImplementedError{"rename not implemented in root storage"}
}

func (s *root) StartChunkedUpload() (string, error) {
	return "", &storage.NotImplementedError{"start chunk upload not implemented in root storage"}
}

func (s *root) PutChunkedObject(identity auth.Identity, r io.Reader, size int64, start int64, chunkID string) error {
	return &storage.NotImplementedError{"put chunked object not implemented in root storage"}
}

func (s *root) CommitChunkedUpload(
	checksum storage.Checksum) error {

	return &storage.NotImplementedError{
		"commit chunked upload not implemented in root storage",
	}
}

func (s *root) Capabilities(identity auth.Identity) storage.Capabilities {
	return &capabilities{}
}
func (s *root) getPathWithoutStoragePrefix(resourcePath string) string {
	parts := strings.Split(resourcePath, "/")
	if len(parts) == 1 {
		return ""
	} else {
		return strings.Join(parts[1:], "/")
	}
}
func (s *root) getPathWithStoragePrefix(relPath string) string {
	return path.Join(s.Prefix(), path.Clean(relPath))
}

type meta struct {
	id          string
	path        string
	size        uint64
	checksum    storage.Checksum
	isContainer bool
	modified    uint64
	etag        string
	mimeType    string
	permissions storage.ResourceMode
	children    []storage.MetaData
}

func (m *meta) ID() string                        { return m.id }
func (m *meta) Path() string                      { return m.path }
func (m *meta) Size() uint64                      { return m.size }
func (m *meta) IsContainer() bool                 { return m.isContainer }
func (m *meta) Modified() uint64                  { return m.modified }
func (m *meta) ETag() string                      { return m.etag }
func (m *meta) MimeType() string                  { return m.mimeType }
func (m *meta) Permissions() storage.ResourceMode { return m.permissions }
func (m *meta) Checksum() storage.Checksum        { return m.checksum }
func (m *meta) Children() []storage.MetaData      { return m.children }
func (m *meta) Extra() interface{}                { return m.Extra() }

type capabilities struct{}

func (c *capabilities) PutObject() bool               { return false }
func (c *capabilities) PutObjectInChunks() bool       { return false }
func (c *capabilities) GetObject() bool               { return false }
func (c *capabilities) GetObjectByByteRange() bool    { return false }
func (c *capabilities) Stat() bool                    { return true }
func (c *capabilities) Remove() bool                  { return false }
func (c *capabilities) CreateContainer() bool         { return false }
func (c *capabilities) Copy() bool                    { return false }
func (c *capabilities) Rename() bool                  { return false }
func (c *capabilities) ThirdPartyCopy() bool          { return false }
func (c *capabilities) ThirdPartyRename() bool        { return false }
func (c *capabilities) ListVersions() bool            { return false }
func (c *capabilities) GetVersion() bool              { return false }
func (c *capabilities) CreateVersion() bool           { return false }
func (c *capabilities) RollbackVersion() bool         { return false }
func (c *capabilities) ListDeletedResources() bool    { return false }
func (c *capabilities) RestoreDeletedResource() bool  { return false }
func (c *capabilities) PurgeDeletedResource() bool    { return false }
func (c *capabilities) VerifyClientChecksum() bool    { return false }
func (c *capabilities) SendChecksum() bool            { return false }
func (c *capabilities) CreateUserHomeDirectory() bool { return false }
func (c *capabilities) SupportedChecksum() string     { return "" }
