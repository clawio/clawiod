// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package local implements a local filesystem.
package local

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	"github.com/clawio/clawiod/pkg/auth"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/storage"
	"github.com/clawio/clawiod/pkg/xattr"
	"hash"
	"hash/adler32"
	"io"
	"mime"
	"os"
	"path"
	"strings"
	"syscall"
)

const (
	DIR_PERM           = 0775
	SUPPORTED_CHECKSUM = "md5"
)

// local is the implementation of the Storage interface to use a local
// filesystem as the storage backend.
type local struct {
	storagePrefix string
	aero          *Aero
	config.Config
	logger.Logger
}

// New creates a local object or returns an error.
func New(storagePrefix string, cfg config.Config,
	log logger.Logger) (storage.Storage, error) {

	aero, err := NewAero(cfg, log)
	if err != nil {
		return nil, err
	}

	s := &local{storagePrefix: storagePrefix, aero: aero,
		Config: cfg, Logger: log}

	return s, nil
}

func (s *local) Prefix() string {
	return s.storagePrefix
}

func (s *local) Capabilities(identity auth.Identity) storage.Capabilities {
	// TOOD: Maybe in the future depending on the user one can give some
	//  capabilities or not. This can be helpful to test new things like
	// allowing some users access to edge features.
	cap := capabilities{}
	return &cap
}

func (s *local) CreateUserHomeDirectory(identity auth.Identity) error {
	return s.createUserHomeDirectory(identity)
}

func (s *local) PutObject(identity auth.Identity, resourcePath string,
	r io.Reader, size int64, checksum storage.Checksum) error {

	return s.putObject(identity, resourcePath, r, size, checksum)

}

func (s *local) Stat(identity auth.Identity, resourcePath string,
	children bool) (storage.MetaData, error) {

	return s.stat(identity, resourcePath, children)
}

func (s *local) GetObject(identity auth.Identity,
	resourcePath string) (io.Reader, error) {

	return s.getObject(identity, resourcePath)
}

func (s *local) Remove(identity auth.Identity, resourcePath string,
	recursive bool) error {

	return s.remove(identity, resourcePath, recursive)
}

func (s *local) CreateContainer(identity auth.Identity,
	resourcePath string) error {

	return s.createContainer(identity, resourcePath)
}

func (s *local) Copy(identity auth.Identity, fromPath, toPath string) error {
	_, fromAbsPath := s.getRelAndAbsPaths(fromPath, identity)
	_, toAbsPath := s.getRelAndAbsPaths(toPath, identity)

	tmpPath := s.getTmpPath()

	s.Info("local: copy " + fromAbsPath + " to " + toAbsPath)

	// Is it a container ?
	meta, err := s.Stat(identity, fromPath, false)
	if err != nil {
		return err
	}

	// If it is an object, just copy it.
	if !meta.IsContainer() {
		err = s.stageFile(fromAbsPath, tmpPath, int64(meta.Size()))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toAbsPath))
	}

	// It is a container, so the copy is recursive.
	err = s.stageDir(fromAbsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toAbsPath))
}

func (s *local) Rename(identity auth.Identity, fromPath, toPath string) error {
	return s.rename(identity, fromPath, toPath)
}

func (s *local) StartChunkedUpload() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *local) PutChunkedObject(identity auth.Identity, r io.Reader,
	size int64, start int64, chunkID string) error {

	return fmt.Errorf("not implemented")
}

func (s *local) CommitChunkedUpload(checksum storage.Checksum) error {

	return fmt.Errorf("not implemented")
}

func (s *local) rename(identity auth.Identity, fromPath, toPath string) error {
	_, fromAbsPath := s.getRelAndAbsPaths(fromPath, identity)
	_, toAbsPath := s.getRelAndAbsPaths(toPath, identity)

	s.Info("local: rename " + fromAbsPath + " to " + toAbsPath)

	return s.convertError(os.Rename(fromAbsPath, toAbsPath))
}
func (s *local) copy(identity auth.Identity, fromPath, toPath string) error {
	_, fromAbsPath := s.getRelAndAbsPaths(fromPath, identity)
	_, toAbsPath := s.getRelAndAbsPaths(toPath, identity)

	tmpPath := s.getTmpPath()

	s.Info("local: copy " + fromAbsPath + " to " + toAbsPath)

	// Is it a container ?
	meta, err := s.Stat(identity, fromPath, false)
	if err != nil {
		return err
	}

	// If it is an object, just copy it.
	if !meta.IsContainer() {
		err = s.stageFile(fromAbsPath, tmpPath, int64(meta.Size()))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toAbsPath))
	}

	// It is a container, so the copy is recursive.
	err = s.stageDir(fromAbsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toAbsPath))
}

func (s *local) getObject(identity auth.Identity,
	resourcePath string) (io.Reader, error) {

	_, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	s.Info("local: get " + absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	return file, nil
}
func (s *local) putObject(identity auth.Identity, resourcePath string,
	r io.Reader, size int64, checksum storage.Checksum) error {

	_, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	s.Info("local: put " + absPath)

	tmpPath := s.getTmpPath()

	fd, err := os.Create(tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				absPath, err.Error())

			s.Warning(msg)
		}
	}()

	var mw io.Writer
	var hasher hash.Hash
	var isChecksumed bool
	var computedChecksum string

	// Select hasher based on capabilities. TODO: add more
	srvChk := s.Capabilities(identity).SupportedChecksum()
	switch srvChk {
	case "md5":
		hasher = md5.New()
		isChecksumed = true
		mw = io.MultiWriter(fd, hasher)
	case "sha1":
		hasher = sha1.New()
		isChecksumed = true
		mw = io.MultiWriter(fd, hasher)
	case "adler32":
		hasher = adler32.New()
		isChecksumed = true
		mw = io.MultiWriter(fd, hasher)
	default:
		mw = io.MultiWriter(fd)
	}

	// Write to tmp file
	_, err = io.CopyN(mw, r, size)
	if err != nil {
		return s.convertError(err)
	}

	if isChecksumed {
		// checksums are given in hexadecimal format.
		computedChecksum = fmt.Sprintf("%x", string(hasher.Sum(nil)))

		if s.Capabilities(identity).VerifyClientChecksum() &&
			checksum.Type() == srvChk && checksum.Value() != "" {

			isCorrupted := computedChecksum != checksum.Value()

			if isCorrupted {
				err := &storage.BadChecksumError{
					Computed: checksum.Type() + ":" + computedChecksum,
					Expected: checksum.String()}

				s.Err(err.Error())
				return s.convertError(err)
			}
		}
		err = xattr.SetXAttr(tmpPath, "user.checksum",
			[]byte(srvChk+":"+computedChecksum), xattr.XAttrCreateOrReplace)

		if err != nil {
			return s.convertError(err)
		}
	}

	resourceID := uuid.New()
	err = xattr.SetXAttr(tmpPath, "id", []byte(resourceID), xattr.XAttrCreate)

	// Atomic move from tmp file to target file.
	err = s.commitPutFile(tmpPath, absPath)
	if err != nil {
		return s.convertError(err)
	}

	// Propagate changes.
	err = s.aero.PutRecord(resourcePath, resourceID)
	if err != nil {
		return err
	}

	return nil
}

func (s *local) remove(identity auth.Identity, resourcePath string,
	recursive bool) error {

	_, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	s.Info("local: remove " + absPath)

	if recursive == false {
		return s.convertError(os.Remove(absPath))
	}
	return s.convertError(os.RemoveAll(absPath))
}

func (s *local) getMergedMetaData(resourcePath string,
	identity auth.Identity) (*meta, error) {

	m, err := s.getFSInfo(resourcePath, identity)
	if err != nil {
		return nil, s.convertError(err)
	}

	rec, err := s.aero.GetOrCreateRecord(resourcePath)
	if err != nil {
		return nil, s.convertError(err)
	}

	m.modified = uint64(rec.Bins["mtime"].(int))
	m.etag = rec.Bins["etag"].(string)
	return m, nil

}

func (s *local) stat(identity auth.Identity, resourcePath string,
	children bool) (storage.MetaData, error) {

	m, err := s.getMergedMetaData(resourcePath, identity)
	if err != nil {
		return nil, s.convertError(err)
	}

	if !m.IsContainer() || children == false {
		return m, nil
	}

	fns, err := s.getFSChildrenNames(resourcePath, identity)
	if err != nil {
		return nil, s.convertError(err)
	}
	panic(fns)
	return m, nil
	/*
		if !m.IsContainer() || children == false {
			return m, nil
		}

		fd, err := os.Open(absPath)
		if err != nil {
			return nil, s.convertError(err)
		}
		defer func() {
			if err := fd.Close(); err != nil {
				msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
					absPath, err.Error())

				s.Warning(msg)
			}
		}()

		finfos, err := fd.Readdir(0)
		if err != nil {
			return nil, s.convertError(err)
		}

		m.children = make([]storage.MetaData, len(finfos))
		for i, f := range finfos {
			childPath := path.Join(m.Path(), path.Clean(f.Name()))
			mimeType := s.getMimeType(f)
			permChild := s.getPermissions(f)
			if f.IsDir() {
				childPath += "/"
			}
			cm := meta{
				id:          childPath,
				path:        childPath,
				size:        uint64(f.Size()),
				isContainer: f.IsDir(),
				modified:    uint64(f.ModTime().UnixNano()),
				etag:        fmt.Sprintf("%d", f.ModTime().UnixNano()),
				mimeType:    mimeType,
				permissions: permChild,
			}
			m.children[i] = &cm
		}
		return m, nil*/
}

func (s *local) getFSInfo(resourcePath string,
	identity auth.Identity) (*meta, error) {

	relPath, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	s.Info("local: stat " + absPath)

	// Get storage file info.
	finfo, err := os.Stat(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}

	id, err := xattr.GetXAttr(absPath, "user.cid")
	if err != nil {
		if err == syscall.ENODATA {
			id = []byte(uuid.New())
			err = xattr.SetXAttr(absPath, "user.cid", []byte(id), xattr.XAttrCreate)
			if err != nil {
				return nil, err
			}
			err := s.aero.PutRecord(resourcePath, string(id))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if len(id) == 0 { // xattr is empty but is set
		id = []byte(uuid.New())
		err = xattr.SetXAttr(absPath, "user.cid", []byte(id), xattr.XAttrCreateOrReplace)
		if err != nil {
			return nil, err
		}
		err := s.aero.PutRecord(resourcePath, string(id))
		if err != nil {
			return nil, err
		}
	}

	mimeType := s.getMimeType(finfo)
	perm := s.getPermissions(finfo)
	parentPath := s.pathWithPrefix(relPath)
	if finfo.IsDir() {
		parentPath += "/" // container' path ends with slash
	}

	m := meta{
		id:          string(id),
		path:        parentPath,
		size:        uint64(finfo.Size()),
		isContainer: finfo.IsDir(),
		mimeType:    mimeType,
		permissions: perm,
	}

	return &m, nil
}

func (s *local) getFSChildrenNames(resourcePath string,
	identity auth.Identity) ([]string, error) {

	_, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	fd, err := os.Open(absPath)
	if err != nil {
		return nil, s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				absPath, err.Error())

			s.Warning(msg)
		}
	}()

	fns, err := fd.Readdirnames(0)
	if err != nil {
		return nil, s.convertError(err)
	}
	return fns, nil
}

func (s *local) createContainer(identity auth.Identity,
	resourcePath string) error {

	_, absPath := s.getRelAndAbsPaths(resourcePath, identity)

	s.Info("local: createcontainer " + absPath)

	err := os.Mkdir(absPath, DIR_PERM)
	if err != nil {
		return s.convertError(err)
	}

	// Set xattrs, on moves they are preserved.
	resourceID := uuid.New()
	err = xattr.SetXAttr(absPath, "id", []byte(resourceID), xattr.XAttrCreate)

	// On Mac, it always gives an error with the error string errno 0.
	//  False positive.
	if err != nil && !strings.Contains(err.Error(), "errno 0") {
		return err
	}

	return s.aero.PutRecord(resourcePath, resourceID)
}

func (s *local) createUserHomeDirectory(identity auth.Identity) error {
	exists, err := s.isHomeDirCreated(identity)
	if err != nil {
		return s.convertError(err)
	}
	if exists {
		return nil
	}
	homeDir := path.Join(s.GetDirectives().LocalStorageRootDataDir,
		path.Join(identity.AuthTypeID(), identity.PID()))

	return s.convertError(os.MkdirAll(homeDir, DIR_PERM))
}
func (s *local) isHomeDirCreated(identity auth.Identity) (bool, error) {
	homeDir := path.Join(s.GetDirectives().LocalStorageRootDataDir,
		path.Join(identity.AuthTypeID(), identity.PID()))

	_, err := os.Stat(homeDir)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
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

func (s *local) getTmpPath() string {
	return path.Join(s.GetDirectives().LocalStorageRootTmpDir, uuid.New())
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
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				source, err.Error())

			s.Warning(msg)
		}
	}()

	writer, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				dest, err.Error())

			s.Warning(msg)
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
	err = os.MkdirAll(dest, DIR_PERM)
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	defer func() {
		if err := directory.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				source, err.Error())

			s.Warning(msg)
		}
	}()

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

func (s *local) sanitizePath(resourcePath string) string {
	return resourcePath
}

func (s *local) pathWithoutPrefix(resourcePath string) string {
	parts := strings.Split(resourcePath, "/")
	if len(parts) == 1 {
		return ""
	} else {
		return strings.Join(parts[1:], "/")
	}
}
func (s *local) pathWithPrefix(relPath string) string {
	return path.Join(s.Prefix(), path.Clean(relPath))
}

func (s *local) getMimeType(fi os.FileInfo) string {
	if fi.IsDir() {
		return storage.DEFAULT_CONTAINER_MIMETYPE
	}
	mimeType := mime.TypeByExtension(path.Ext(fi.Name()))
	if mimeType == "" {
		mimeType = storage.DEFAULT_OBJECT_MIMETYPE
	}
	return mimeType
}

func (s *local) getPermissions(fi os.FileInfo) storage.ResourceMode {
	perm := storage.PSTAT | storage.PRM
	if fi.IsDir() {
		return perm | storage.PLIST
	}
	return perm | storage.PGET
}

// getRelAndAbsPaths returns the relativePath (without storage prefix)
// and the absolutePath (the fs path)
func (s *local) getRelAndAbsPaths(resourcePath string,
	identity auth.Identity) (string, string) {

	relPath := s.pathWithoutPrefix(resourcePath)
	absPath := path.Join(s.GetDirectives().LocalStorageRootDataDir,
		path.Join(identity.AuthTypeID(), identity.PID(), relPath))

	return relPath, absPath
}

// meta represents the metadata associated with a resources.
// It the fusion of the storageInfo and the hyperInfo.
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
	extra       interface{}
}

/*
func newMeta(si *storageInfo, hi *hyperInfo) *meta {
	m := &meta{}
	m.id = hi.ID
	m.path = storageInfo.ResourcePath
	m.size = uint64(si.Size())
	m.checksum = hi.Checksum
	m.isContainer = si.IsDir()
	m.modified = hi.Mtime
	m.etag = hi.Etag
	m.mimeType = storageInfo.MimeType
	m.permissions = si.Permissions
}
*/
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
func (m *meta) Extra() interface{}                { return m.extra }

type capabilities struct{}

func (c *capabilities) PutObject() bool               { return false }
func (c *capabilities) PutObjectInChunks() bool       { return false }
func (c *capabilities) GetObject() bool               { return false }
func (c *capabilities) GetObjectByByteRange() bool    { return false }
func (c *capabilities) Stat() bool                    { return false }
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
func (c *capabilities) VerifyClientChecksum() bool    { return true }
func (c *capabilities) SendChecksum() bool            { return false }
func (c *capabilities) CreateUserHomeDirectory() bool { return false }
func (c *capabilities) SupportedChecksum() string {
	return SUPPORTED_CHECKSUM
}

// storageInfo represents the information obtainable local filesystem.
type storageInfo struct {
	os.FileInfo
	MimeType     string
	Permissions  storage.ResourceMode
	ResourcePath string
}
