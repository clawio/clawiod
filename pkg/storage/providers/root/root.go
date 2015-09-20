// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package root implements the storage interface to offer a pseude-storage that provides a view to all of the other storages like root or eos.
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
	cfg           config.Config
	log           logger.Logger
}

// New creates a root object or returns an error.\
// to have a
func New(storagePrefix string, storages []storage.Storage, cfg config.Config, log logger.Logger) storage.Storage {
	s := &root{storagePrefix: storagePrefix, storages: storages, cfg: cfg, log: log}
	return s
}

func (s *root) GetStoragePrefix() string {
	return s.storagePrefix
}

func (s *root) CreateUserHomeDirectory(identity *auth.Identity) error {
	return &storage.NotImplementedError{"create user home directory not implemented in root storage"}
}

func (s *root) PutObject(identity *auth.Identity, resourcePath string, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error {
	return &storage.NotImplementedError{"put object not implemented in root storage"}
}
func (s *root) GetObject(identity *auth.Identity, resourcePath string) (io.Reader, error) {
	return nil, &storage.NotImplementedError{"get object not implemented in root storage"}
}
func (s *root) Stat(identity *auth.Identity, resourcePath string, children bool) (*storage.MetaData, error) {

	ts := time.Now().Unix()
	tsString := strconv.FormatInt(ts, 10)

	parentMeta := &storage.MetaData{}
	parentMeta.ID = s.GetStoragePrefix()
	parentMeta.Path = s.GetStoragePrefix()
	parentMeta.IsContainer = true
	parentMeta.MimeType = "inode/container"
	parentMeta.Modified = uint64(ts)
	parentMeta.ETag = tsString

	if children == false {
		return parentMeta, nil
	}

	parentMeta.Children = []*storage.MetaData{}
	for _, sto := range s.storages {
		m := &storage.MetaData{}
		m.ID = sto.GetStoragePrefix()
		m.Path = m.ID
		m.IsContainer = true
		m.MimeType = "inode/container"

		m.Modified = uint64(ts)
		m.ETag = tsString
		parentMeta.Children = append(parentMeta.Children, m)
	}

	return parentMeta, nil
}

func (s *root) Remove(identity *auth.Identity, resourcePath string, recursive bool) error {
	return &storage.NotImplementedError{"remove not implemented in root storage"}
}

func (s *root) CreateContainer(identity *auth.Identity, resourcePath string, recursive bool) error {
	return &storage.NotImplementedError{"create container not implemented in root storage"}
}

func (s *root) Copy(identity *auth.Identity, fromPath, toPath string) error {
	return &storage.NotImplementedError{"copy not implemented in root storage"}
}

func (s *root) Rename(identity *auth.Identity, fromPath, toPath string) error {
	return &storage.NotImplementedError{"rename not implemented in root storage"}
}

func (s *root) StartChunkedUpload() (string, error) {
	return "", &storage.NotImplementedError{"start chunk upload not implemented in root storage"}
}

func (s *root) PutChunkedObject(identity *auth.Identity, r io.Reader, size int64, start int64, chunkID string) error {
	return &storage.NotImplementedError{"put chunked object not implemented in root storage"}
}

func (s *root) CommitChunkedUpload(chunkID string, verifyChecksum bool, checksum, checksumType string) error {
	return &storage.NotImplementedError{"commit chunked upload not implemented in root storage"}
}

func (s *root) GetCapabilities(identity *auth.Identity) *storage.Capabilities {
	// Root storage ONLY has list capability
	cap := storage.Capabilities{}
	cap.Stat = true
	return &cap
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
	return path.Join(s.GetStoragePrefix(), path.Clean(relPath))
}
