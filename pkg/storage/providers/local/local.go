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
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
)

// local is the implementation of the StorageProvider interface to use a local
// filesystem as the storage backend.
type local struct {
	scheme string
	cfg    *config.Config
	log    logger.Logger
}

// New creates a local object or returns an error.
func New(scheme string, cfg *config.Config, log logger.Logger) storage.Storage {
	s := &local{scheme: scheme, cfg: cfg, log: log}
	return s
}

func (s *local) GetScheme() string {
	return s.scheme
}

func (s *local) CreateUserHome(identity *auth.Identity) error {
	exists, err := s.IsUserHomeCreated(identity)
	if err != nil {
		return s.convertError(err)
	}
	if exists {
		return nil
	}
	homeDir := filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID)
	return s.convertError(os.MkdirAll(homeDir, 0666))
}

func (s *local) IsUserHomeCreated(identity *auth.Identity) (bool, error) {
	homeDir := filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID)
	_, err := os.Stat(homeDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *local) PutFile(identity *auth.Identity, uri *url.URL, r io.Reader, size int64, verifyChecksum bool, checksumType, checksum string) error {
	s.cleanURI(uri)
	absPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
	tmpPath := filepath.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir, filepath.Base(uri.Path)+"-"+s.log.RID())

	// We will have the file in the tmp folder so we can calculate checksums
	if verifyChecksum == true && checksumType != "" { // we allow puting file is the checksum type is empty
		if s.isChecksumTypeSupported(checksumType) == false {
			return &storage.UnsupportedChecksumTypeError{Err: fmt.Sprintf("checksum type '%s' not supported", checksumType)}
		}
		fd, err := os.Create(tmpPath)
		if err != nil {
			return s.convertError(err)
		}
		defer func() {
			if err := fd.Close(); err != nil {
				s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": uri.String(), "err": err})
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
				"auth_id":       identity.AuthID,
				"username":      identity.ID,
				"storage":       s.GetScheme(),
				"resource":      absPath,
				"checksum_type": checksumType,
				"expected":      checksum,
				"computed":      computedChecksumHex,
			})
			return &storage.BadChecksumError{Err: fmt.Sprintf("expected '%s' but computed '%s'", checksum, computedChecksumHex)}
		}
		return s.commitPutFile(tmpPath, absPath)
	}
	fd, err := os.Create(tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": uri.String(), "err": err})
		}
	}()
	_, err = io.CopyN(fd, r, size)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(s.commitPutFile(tmpPath, absPath))

}

func (s *local) Stat(identity *auth.Identity, uri *url.URL, children bool) (*storage.MetaData, error) {
	var finfo os.FileInfo
	var absPath string
	/*if id != "" {
		absPath = filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, id))
		fi, err := os.Stat(absPath)
		if err != nil {
			return nil, s.convertError(err)
		}
		finfo = fi
		uri = &url.URL{}
		uri.Scheme = s.GetScheme()
		uri.Path = id
	} else {
		if uri.Path == "" {
			return nil, &storage.NotExistError{Err: fmt.Sprintln("No path provided")}
		}
		s.cleanURI(uri)
		absPath = filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
		fi, err := os.Stat(absPath)
		if err != nil {
			return nil, s.convertError(err)
		}
		finfo = fi
	}*/
	if uri.Path == "" {
		return nil, &storage.NotExistError{Err: fmt.Sprintln("No path provided")}
	}
	s.cleanURI(uri)
	absPath = filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
	fi, err := os.Stat(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	finfo = fi

	mimeType := mime.TypeByExtension(filepath.Ext(uri.Path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if finfo.IsDir() {
		mimeType = "inode/directory"
	}
	meta := storage.MetaData{
		ID:       uri.String(),
		Path:     uri.String(),
		Size:     uint64(finfo.Size()),
		IsCol:    finfo.IsDir(),
		Modified: uint64(finfo.ModTime().Unix()),
		ETag:     fmt.Sprintf("\"%d\"", finfo.ModTime().Unix()),
		MimeType: mimeType,
	}

	if meta.IsCol == false {
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
			s.log.Warningf("Cannot close resource: %+v", map[string]interface{}{"resource": uri.String(), "err": err})
		}
	}()

	finfos, err := fd.Readdir(0)
	if err != nil {
		return nil, s.convertError(err)
	}

	meta.Children = make([]*storage.MetaData, len(finfos))
	for i, f := range finfos {
		childPathURI := *uri
		childPathURI.Path = filepath.Join(childPathURI.Path, f.Name())
		childPath := childPathURI.String()

		uri.Fragment = ""
		uri.RawQuery = ""
		mimeType := mime.TypeByExtension(filepath.Ext(childPath))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		if f.IsDir() {
			mimeType = "inode/directory"
		}
		m := storage.MetaData{
			ID:       childPath,
			Path:     childPath,
			Size:     uint64(f.Size()),
			IsCol:    f.IsDir(),
			Modified: uint64(f.ModTime().Unix()),
			ETag:     fmt.Sprintf("\"%d\"", f.ModTime().Unix()),
			MimeType: mimeType,
		}
		meta.Children[i] = &m
	}
	s.log.Debugf("Meta: %+v", meta)
	return &meta, nil
}

func (s *local) GetFile(identity *auth.Identity, uri *url.URL) (io.Reader, error) {
	/*if id != "" {
		absPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, id))
		file, err := os.Open(absPath)
		if err != nil {
			return nil, s.convertError(err)
		}
		return file, nil
	}*/
	s.cleanURI(uri)
	absPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
	file, err := os.Open(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	return file, nil
}

func (s *local) Remove(identity *auth.Identity, uri *url.URL, recursive bool) error {
	s.cleanURI(uri)
	absPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
	if recursive == false {
		return s.convertError(os.Remove(absPath))
	}
	return s.convertError(os.RemoveAll(absPath))
}

func (s *local) CreateCol(identity *auth.Identity, uri *url.URL, recursive bool) error {
	s.cleanURI(uri)
	absPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, uri.Path))
	if recursive == false {
		return s.convertError(os.Mkdir(absPath, 0666))
	}
	return s.convertError(os.MkdirAll(absPath, 0666))
}

func (s *local) Copy(identity *auth.Identity, fromURI, toURI *url.URL) error {
	s.cleanURI(fromURI)
	s.cleanURI(toURI)
	fromabsPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, fromURI.Path))
	toabsPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, toURI.Path))
	tmpPath := filepath.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir, filepath.Base(fromURI.Path)+"-"+s.log.RID())

	// we need to get metadata to check if it is a col or file
	meta, err := s.Stat(identity, fromURI, false)
	if err != nil {
		return err
	}

	// we copy the file
	if meta.IsCol == false {
		err = s.stageFile(fromabsPath, tmpPath, int64(meta.Size))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toabsPath))
	}

	err = s.stageDir(fromabsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toabsPath))
}

func (s *local) Rename(identity *auth.Identity, fromURI, toURI *url.URL) error {
	s.cleanURI(fromURI)
	s.cleanURI(toURI)
	fromabsPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, fromURI.Path))
	toabsPath := filepath.Clean(filepath.Join(s.cfg.GetDirectives().LocalStorageRootDataDir, identity.AuthID, identity.ID, toURI.Path))
	return s.convertError(os.Rename(fromabsPath, toabsPath))
}

func (s *local) convertError(err error) error {
	if err == nil {
		return nil
	} else if os.IsExist(err) {
		return &storage.ExistError{Err: err.Error()}
	} else if os.IsNotExist(err) {
		return &storage.NotExistError{Err: err.Error()}
	}
	return err
}

func (s *local) GetCapabilities() *storage.Capabilities {
	cap := storage.Capabilities{}
	return &cap
}

func (s *local) GetSupportedChecksumTypes() []string {
	return s.cfg.GetDirectives().LocalStorageSupportedChecksumTypes
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

// clean URI removes unncessary information about the URI like User, Host, Fragment and Query.
// The local stoage does not need them so we can safely remove them to not use
// them in stat responses.
func (s *local) cleanURI(uri *url.URL) {
	uri.User = nil
	uri.Host = ""
	uri.RawQuery = ""
	uri.Fragment = ""
}

func (s *local) isChecksumTypeSupported(checksumType string) bool {
	for _, cs := range s.cfg.GetDirectives().LocalStorageSupportedChecksumTypes {
		if cs == checksumType {
			return true
		}
	}
	return false
}
