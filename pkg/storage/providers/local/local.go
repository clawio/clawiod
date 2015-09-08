// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package local implements the storage interface to use a local filesystem as a storage backend.
package local

import (
	"crypto/md5"
	"fmt"
	"io"
	"mime"
	"os"
	"path"
	"strings"

	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
)

// local is the implementation of the Storage interface to use a local
// filesystem as the storage backend.
type local struct {
	storagePrefix string
	cfg           *config.Config
	log           logger.Logger
}

// New creates a local object or returns an error.
func New(storagePrefix string, cfg *config.Config, log logger.Logger) storage.Storage {
	s := &local{storagePrefix: storagePrefix, cfg: cfg, log: log}
	return s
}

func (s *local) GetStoragePrefix() string {
	return s.storagePrefix
}

func (s *local) CreateUserHomeDirectory(identity *auth.Identity) error {
	exists, err := s.isUserHomeDirectoryCreated(identity)
	if err != nil {
		return s.convertError(err)
	}
	if exists {
		return nil
	}
	homeDir := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN))
	return s.convertError(os.MkdirAll(homeDir, 0666))
}

func (s *local) PutObject(identity *auth.Identity, resourcePath string, r io.Reader, size int64, verifyChecksum bool, checksum, checksumType string) error {
	relPath := s.getPathWithoutStoragePrefix(resourcePath)
	absPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, relPath))
	tmpPath := path.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir, path.Join(path.Base(relPath), "-", s.log.RID()))

	// If the checksum type is the same as the one in the storage capabilities object, then we do it.
	if verifyChecksum == true && checksumType == s.GetCapabilities().SupportedChecksum {
		fd, err := os.Create(tmpPath)
		if err != nil {
			return s.convertError(err)
		}
		defer func() {
			if err := fd.Close(); err != nil {
				s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": absPath, "err": err})
			}
		}()

		hasher := md5.New()
		mw := io.MultiWriter(fd, hasher)
		_, err = io.CopyN(mw, r, size)
		if err != nil {
			return s.convertError(err)
		}
		computedChecksum := string(hasher.Sum(nil))
		computedChecksumHex := fmt.Sprintf("%x", computedChecksum)
		if computedChecksumHex != checksum {
			s.log.Errf("Data corruption: %+v", map[string]interface{}{
				"authid":       identity.AuthID,
				"id":           identity.EPPN,
				"storage":      s.GetStoragePrefix(),
				"resource":     absPath,
				"checksumtype": checksumType,
				"expected":     checksum,
				"computed":     computedChecksumHex,
			})
			return &storage.BadChecksumError{Err: fmt.Sprintf("expected:%s but computed:%s", checksum, computedChecksumHex)}
		}
		return s.commitPutFile(tmpPath, absPath)
	}

	fd, err := os.Create(tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": absPath, "err": err})
		}
	}()
	_, err = io.CopyN(fd, r, size)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(s.commitPutFile(tmpPath, absPath))

}

func (s *local) Stat(identity *auth.Identity, resourcePath string, children bool) (*storage.MetaData, error) {
	relPath := s.getPathWithoutStoragePrefix(resourcePath)

	absPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, relPath))
	finfo, err := os.Stat(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}

	mimeType := mime.TypeByExtension(path.Ext(relPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if finfo.IsDir() {
		mimeType = "inode/container"
	}
	meta := storage.MetaData{
		ID:          s.getPathWithStoragePrefix(relPath),
		Path:        s.getPathWithStoragePrefix(relPath),
		Size:        uint64(finfo.Size()),
		IsContainer: finfo.IsDir(),
		Modified:    uint64(finfo.ModTime().Unix()),
		ETag:        fmt.Sprintf("\"%d\"", finfo.ModTime().Unix()),
		MimeType:    mimeType,
	}

	if !meta.IsContainer {
		return &meta, nil
	}
	if children == false {
		return &meta, nil
	}

	fd, err := os.Open(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": relPath, "err": err})
		}
	}()

	finfos, err := fd.Readdir(0)
	if err != nil {
		return nil, s.convertError(err)
	}

	meta.Children = make([]*storage.MetaData, len(finfos))
	for i, f := range finfos {
		childPath := path.Join(relPath, path.Clean(f.Name()))
		mimeType := mime.TypeByExtension(path.Ext(childPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		if f.IsDir() {
			mimeType = "inode/container"
		}
		m := storage.MetaData{
			ID:          s.getPathWithStoragePrefix(childPath),
			Path:        s.getPathWithStoragePrefix(childPath),
			Size:        uint64(f.Size()),
			IsContainer: f.IsDir(),
			Modified:    uint64(f.ModTime().Unix()),
			ETag:        fmt.Sprintf("\"%d\"", f.ModTime().Unix()),
			MimeType:    mimeType,
		}
		meta.Children[i] = &m
	}
	s.log.Debugf("Meta: %+v", meta)
	return &meta, nil
}

func (s *local) GetObject(identity *auth.Identity, resourcePath string) (io.Reader, error) {
	relPath := s.getPathWithoutStoragePrefix(resourcePath)
	absPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, relPath))
	file, err := os.Open(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	return file, nil
}

func (s *local) Remove(identity *auth.Identity, resourcePath string, recursive bool) error {
	relPath := s.getPathWithoutStoragePrefix(resourcePath)
	absPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, relPath))
	if recursive == false {
		return s.convertError(os.Remove(absPath))
	}
	return s.convertError(os.RemoveAll(absPath))
}

func (s *local) CreateContainer(identity *auth.Identity, resourcePath string, recursive bool) error {
	relPath := s.getPathWithoutStoragePrefix(resourcePath)
	absPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, relPath))
	if recursive == false {
		return s.convertError(os.Mkdir(absPath, 0666))
	}
	return s.convertError(os.MkdirAll(absPath, 0666))
}

func (s *local) Copy(identity *auth.Identity, fromPath, toPath string) error {
	fromRelPath := s.getPathWithoutStoragePrefix(fromPath)
	toRelPath := s.getPathWithoutStoragePrefix(toPath)
	fromAbsPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, fromRelPath))
	toAbsPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, toRelPath))
	tmpPath := path.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir, path.Join(path.Base(fromRelPath), "-", s.log.RID()))

	// we need to get metadata to check if it is a col or file
	meta, err := s.Stat(identity, fromPath, false)
	if err != nil {
		return err
	}

	// we copy the file
	if !meta.IsContainer {
		err = s.stageFile(fromAbsPath, tmpPath, int64(meta.Size))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toAbsPath))
	}

	err = s.stageDir(fromAbsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toAbsPath))
}

func (s *local) Rename(identity *auth.Identity, fromPath, toPath string) error {
	fromRelPath := s.getPathWithoutStoragePrefix(fromPath)
	toRelPath := s.getPathWithoutStoragePrefix(toPath)
	fromAbsPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, fromRelPath))
	toAbsPath := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN, toRelPath))
	return s.convertError(os.Rename(fromAbsPath, toAbsPath))
}

func (s *local) StartChunkedUpload() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *local) PutChunkedObject(identity *auth.Identity, r io.Reader, size int64, start int64, chunkID string) error {
	return fmt.Errorf("not implemented")
}

func (s *local) CommitChunkedUpload(chunkID string, verifyChecksum bool, checksum, checksumType string) error {
	return fmt.Errorf("not implemented")
}

func (s *local) convertError(err error) error {
	if err == nil {
		return nil
	} else if os.IsExist(err) {
		return &storage.AlreadyExistError{Err: err.Error()}
	} else if os.IsNotExist(err) {
		return &storage.NotExistError{Err: err.Error()}
	}
	return err
}

func (s *local) GetCapabilities() *storage.Capabilities {
	cap := storage.Capabilities{}
	cap.CreateUserHomeDirectory = true
	return &cap
}

func (s *local) commitPutFile(from, to string) error {
	return os.Rename(from, to)
}

func (s *local) stageFile(source string, dest string, size int64) (err error) {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": source, "err": err})
		}
	}()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": dest, "err": err})
		}
	}()

	_, err = io.CopyN(writer, reader, size)
	if err != nil {
		return err
	}
	return nil
}

func (s *local) stageDir(source string, dest string) (err error) {
	// create dest dir
	err = os.MkdirAll(dest, 0644)
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := path.Join(source, obj.Name())
		destinationfilepointer := path.Join(dest, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = s.stageDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = s.stageFile(sourcefilepointer, destinationfilepointer, obj.Size())
			if err != nil {
				return err
			}
		}

	}
	return
}

func (s *local) isUserHomeDirectoryCreated(identity *auth.Identity) (bool, error) {
	homeDir := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, path.Join(identity.AuthID, identity.EPPN))
	_, err := os.Stat(homeDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *local) sanitizePath(resourcePath string) string {
	return resourcePath
}

func (s *local) getPathWithoutStoragePrefix(resourcePath string) string {
	parts := strings.Split(resourcePath, "/")
	if len(parts) == 1 {
		return ""
	} else {
		return strings.Join(parts[1:], "/")
	}
}
func (s *local) getPathWithStoragePrefix(relPath string) string {
	return path.Join(s.GetStoragePrefix(), path.Clean(relPath))
}
