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
	"github.com/clawio/clawiod/Godeps/_workspace/src/golang.org/x/net/context"
	auth "github.com/clawio/clawiod/pkg/auth"
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
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

const (
	DirPerm           = 0775
	SupportedChecksum = "md5"
	XAttrID           = "user.cid"
	XAttrChecksum     = "user.checksum"
)

const ()

// local is the implementation of the Storage interface to use a local
// filesystem as the storage backend.
type local struct {
	prefix string
	aero   *Aero
	cfg    config.Config
}

type NewParams struct {
	Prefix string
	Config config.Config
}

// New creates a local object or returns an error.
func New(p *NewParams) (storage.Storage, error) {

	aero, err := NewAero(p.Config)
	if err != nil {
		return nil, err
	}

	s := &local{prefix: p.Prefix, aero: aero,
		cfg: p.Config}

	return s, nil
}

func (s *local) Prefix() string {
	return s.prefix
}

func (s *local) Capabilities(ctx context.Context, p *storage.CapabilitiesParams) *storage.Capabilities {
	// TOOD: Maybe in the future depending on the user one can give some
	//  capabilities or not. This can be helpful to test new things like
	// allowing some users access to edge features.
	cap := storage.Capabilities{}
	cap.Copy = true
	cap.CreateContainer = true
	cap.CreateUserHomeDir = true
	cap.GetObject = true
	cap.GetObjectByByteRange = true
	cap.PutObject = true
	cap.Remove = true
	cap.Rename = true
	cap.Stat = true
	cap.SupportedChecksum = "md5"
	cap.VerifyClientChecksum = true
	return &cap
}

func (s *local) CreateUserHomeDir(ctx context.Context, p *storage.CreateUserHomeDirParams) error {
	return s.createUserHomeDirectory(ctx, p)
}

func (s *local) PutObject(ctx context.Context, p *storage.PutObjectParams) error {
	// decide where it is a OC chunk upload or not
	// now we decide based on path. Header OC-Chunked = 1 could be used also.
	chunked, err := IsChunked(p.Rsp)
	if err != nil {
		return s.convertError(err)
	}
	if chunked {
		return s.putOCChunkObject(ctx, p)
	}
	return s.putObject(ctx, p)
}

func (s *local) Stat(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {
	return s.stat(ctx, p)
}

func (s *local) GetObject(ctx context.Context, p *storage.GetObjectParams) (io.Reader, error) {
	return s.getObject(ctx, p)
}

func (s *local) Remove(ctx context.Context, p *storage.RemoveParams) error {
	return s.remove(ctx, p)
}

func (s *local) CreateContainer(ctx context.Context, p *storage.CreateContainerParams) error {
	return s.createContainer(ctx, p)
}

func (s *local) Copy(ctx context.Context, p *storage.CopyParams) error {
	log := logger.MustFromContext(ctx)

	_, fromAbsPath := s.getRelAndAbsPaths(p.Src, p.Idt)
	_, toAbsPath := s.getRelAndAbsPaths(p.Dst, p.Idt)

	tmpPath := s.getTmpPath()

	log.Info("local: copy " + fromAbsPath + " to " + toAbsPath)

	// Is it a container ?
	statParams := &storage.StatParams{}
	statParams.BaseParams = p.BaseParams
	statParams.Rsp = p.Src

	meta, err := s.Stat(ctx, statParams)
	if err != nil {
		return err
	}

	// If it is an object, just copy it.
	if !meta.IsContainer {
		err = s.stageFile(ctx, fromAbsPath, tmpPath, int64(meta.Size))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toAbsPath))
	}

	// It is a container, so the copy is recursive.
	err = s.stageDir(ctx, fromAbsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toAbsPath))
}

func (s *local) Rename(ctx context.Context, p *storage.RenameParams) error {
	return s.rename(ctx, p)
}

func (s *local) StartChunkedUpload(ctx context.Context, p *storage.StartChunkUploadParams) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *local) PutChunkedObject(ctx context.Context, p *storage.PutChunkedObjectParams) error {
	return fmt.Errorf("not implemented")
}

func (s *local) CommitChunkedUpload(ctx context.Context, p *storage.CommitChunkUploadParams) error {
	return fmt.Errorf("not implemented")
}

func (s *local) rename(ctx context.Context, p *storage.RenameParams) error {
	log := logger.MustFromContext(ctx)

	_, fromAbsPath := s.getRelAndAbsPaths(p.Src, p.Idt)
	_, toAbsPath := s.getRelAndAbsPaths(p.Dst, p.Idt)

	log.Info("local: rename " + fromAbsPath + " to " + toAbsPath)

	return s.convertError(os.Rename(fromAbsPath, toAbsPath))
}
func (s *local) copy(ctx context.Context, p *storage.CopyParams) error {
	log := logger.MustFromContext(ctx)

	_, fromAbsPath := s.getRelAndAbsPaths(p.Src, p.Idt)
	_, toAbsPath := s.getRelAndAbsPaths(p.Dst, p.Idt)

	tmpPath := s.getTmpPath()

	log.Info("local: copy " + fromAbsPath + " to " + toAbsPath)

	// Is it a container ?
	statParams := &storage.StatParams{}
	statParams.BaseParams = p.BaseParams
	statParams.Rsp = fromAbsPath

	meta, err := s.Stat(ctx, statParams)
	if err != nil {
		return err
	}

	// If it is an object, just copy it.
	if !meta.IsContainer {
		err = s.stageFile(ctx, fromAbsPath, tmpPath, int64(meta.Size))
		if err != nil {
			return s.convertError(err)
		}
		return s.convertError(os.Rename(tmpPath, toAbsPath))
	}

	// It is a container, so the copy is recursive.
	err = s.stageDir(ctx, fromAbsPath, tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	return s.convertError(os.Rename(tmpPath, toAbsPath))
}

func (s *local) getObject(ctx context.Context, p *storage.GetObjectParams) (io.Reader, error) {
	log := logger.MustFromContext(ctx)

	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	log.Info("local: get " + ap)

	file, err := os.Open(ap)
	if err != nil {
		return nil, s.convertError(err)
	}
	if p.Range == nil {
		return file, nil
	}
	_, err = file.Seek(int64(p.Range.Offset), 0)
	if err != nil {
		return nil, err
	}
	return io.LimitReader(file, int64(p.Size)), nil
}
func (s *local) putObject(ctx context.Context, p *storage.PutObjectParams) error {
	log := logger.MustFromContext(ctx)

	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	log.Info("local: put " + ap)

	tmpPath := s.getTmpPath()

	fd, err := os.Create(tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				ap, err.Error())

			log.Warning(msg)
		}
	}()

	var mw io.Writer
	var hasher hash.Hash
	var isChecksumed bool
	var computedChecksum string

	// Select hasher based on capabilities. TODO: add more
	capParams := &storage.CapabilitiesParams{}
	capParams.BaseParams = p.BaseParams
	srvChk := s.Capabilities(ctx, capParams).SupportedChecksum
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
	_, err = io.CopyN(mw, p.Reader, int64(p.Size))
	if err != nil {
		return s.convertError(err)
	}

	if isChecksumed {
		// checksums are given in hexadecimal format.
		computedChecksum = fmt.Sprintf("%x", string(hasher.Sum(nil)))

		if s.Capabilities(ctx, capParams).VerifyClientChecksum &&
			p.Checksum.Type() == srvChk && p.Checksum.Value() != "" {

			isCorrupted := computedChecksum != p.Checksum.Value()

			if isCorrupted {
				err := &storage.BadChecksumError{
					Computed: p.Checksum.Type() + ":" + computedChecksum,
					Expected: p.Checksum.String()}

				log.Err(err.Error())
				return s.convertError(err)
			}
		}
		err = xattr.SetXAttr(tmpPath, XAttrChecksum,
			[]byte(srvChk+":"+computedChecksum), xattr.XAttrCreateOrReplace)

		if err != nil {
			return s.convertError(err)
		}
	}

	resourceID := uuid.New()
	err = xattr.SetXAttr(tmpPath, XAttrID, []byte(resourceID), xattr.XAttrCreate)

	// Atomic move from tmp file to target file.
	err = s.commitPutFile(tmpPath, ap)
	if err != nil {
		return s.convertError(err)
	}

	// Propagate changes.
	err = s.aero.PutRecord(p.Rsp, resourceID)
	if err != nil {
		return err
	}

	return nil
}

func (s *local) putOCChunkObject(ctx context.Context, p *storage.PutObjectParams) error {
	// TODO(labkode) Check if r.ContentLength should be changed by
	// chunkInfo.OCChunkSize in chunk uploads
	log := logger.MustFromContext(ctx)

	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	// cast to ChunkInfo
	//chunkInfo, ok := extra.(*ChunkHeaderInfo)
	//if !ok {
	//	return fmt.Errorf("local: chunk upload without header chunk info")
	//}
	chunkPathInfo, err := GetChunkPathInfo(p.Rsp)
	if err != nil {
		return s.convertError(err)
	}
	log.Info("putOCChunkObject: getted " + chunkPathInfo.String())
	tmpPath := s.getTmpPath()

	fd, err := os.Create(tmpPath)
	if err != nil {
		return s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				ap, err.Error())

			log.Warning(msg)
		}
	}()
	log.Info("putOCChunkObject: created tmpPath for chunk at " + tmpPath)

	/* TODO(labkode) Configurable checksuming of individual chunks
	var mw io.Writer
	var hasher hash.Hash
	var isChecksumed bool
	var computedChecksum string

	// Select hasher based on capabilities. TODO: add more
	srvChk := s.Capabilities(idt).SupportedChecksum()
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

		if s.Capabilities(idt).VerifyClientChecksum() &&
			checksum.Type() == srvChk && checksum.Value() != "" {

			isCorrupted := computedChecksum != checksum.Value()

			if isCorrupted {
				err := &storage.BadChecksumError{
					Computed: checksum.Type() + ":" + computedChecksum,
					Expected: checksum.String()}

				log.Err(err.Error())
				return s.convertError(err)
			}
		}
		err = xattr.SetXAttr(tmpPath, XAttrChecksum,
			[]byte(srvChk+":"+computedChecksum), xattr.XAttrCreateOrReplace)

		if err != nil {
			return s.convertError(err)
		}
	}
	*/

	_, err = io.CopyN(fd, p.Reader, int64(p.Size))
	if err != nil {
		return s.convertError(err)
	}

	log.Debug("putOCChunkObject: copied r.Body to " + tmpPath)

	// At this point the chunk is in the tmp folder.
	// The chunk folder has to be created

	chunkFolder, err := s.getChunkFolder(chunkPathInfo)
	if err != nil {
		return s.convertError(err)
	}

	log.Debug("putOCChunkObject: created chunkFolder at " + chunkFolder)

	chunkDst := path.Join(
		chunkFolder,
		path.Clean(strconv.FormatUint(chunkPathInfo.CurrentChunk, 10)))

	err = os.Rename(tmpPath, chunkDst)

	if err != nil {
		return s.convertError(err)
	}

	log.Debug("putOCChunkObject: moved chunk from " + tmpPath + " to " + chunkDst)

	// Check that all chunks are uploaded.
	// This is very inefficient, the server has to check that it has all the
	// chunks after each uploaded chunk.
	// A two-phase upload like DropBox is better, because the server will
	// assembly the chunks when the client asks for it.

	fdChunkFolder, err := os.Open(chunkFolder)
	if err != nil {
		return s.convertError(err)
	}
	defer fdChunkFolder.Close()
	log.Info("putOCChunkObject: open " + chunkFolder)

	fns, err := fdChunkFolder.Readdirnames(-1)
	if err != nil {
		return s.convertError(err)
	}

	log.Info(fmt.Sprintf("putOCChunkObject: %d out of %d chunks", len(fns),
		chunkPathInfo.TotalChunks))

	if len(fns) < int(chunkPathInfo.TotalChunks) {
		return nil
	}

	// here len(fns) is >= total chunks.
	// if there are more chunks that the ones in chunkPathInfo.TotalChunks
	// means that the client sent the wrong chunk number.
	// When reconstructing, iterate sequentially over all chunks
	// from 0 to chunkPathInfo.chunkPathInfo

	tp := s.getTmpPath()
	//fdAssembly, err := os.OpenFile(tp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fdAssembly, err := os.OpenFile(tp, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return s.convertError(err)
	}
	defer fdAssembly.Close()

	log.Info("putOCChunkObject: opened tmp file assembly at " + tp)

	for chunk := 0; chunk < int(chunkPathInfo.TotalChunks); chunk++ {
		cp := path.Join(chunkFolder,
			strconv.FormatInt(int64(chunk), 10))

		fdChunk, err := os.Open(cp)
		if err != nil {
			return s.convertError(err)
		}
		log.Info("putOCChunkObject: opened chunk file at " + cp)

		_, err = io.Copy(fdAssembly, fdChunk)
		if err != nil {
			return s.convertError(err)
		}
		log.Info("putOCChunkObject: copied from " + cp + " to " + tp)
	}

	resourceID := uuid.New()
	err = xattr.SetXAttr(tp, XAttrID, []byte(resourceID), xattr.XAttrCreate)
	if err != nil {
		return s.convertError(err)
	}

	log.Info(fmt.Sprintf("putOCChunkObject: setted xattr %s to %s at %s",
		XAttrID, resourceID, tp))

	// TODO(labkode) Check that the assembled file size is == to OC-Total-Length
	// TODO(labkode) Compute checksum and put it in xattrs
	// Atomic move from tmp file to target file after all chunks are uploaded.
	_, dst := s.getRelAndAbsPaths(chunkPathInfo.ResourcePath, p.Idt)

	err = s.commitPutFile(tp, dst)
	if err != nil {
		return s.convertError(err)
	}

	log.Info(fmt.Sprintf("putOCChunkObject: moved %s to %s",
		tp, dst))

	// Propagate changes.
	err = s.aero.PutRecord(dst, resourceID)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("putOCChunkObject: assigned resourceID:%s to %s", resourceID, chunkPathInfo.Rsp))

	return nil
}

func (s *local) getChunkFolder(i *chunkPathInfo) (string, error) {
	// not using the resource path in the chunk folder name allows uploading
	// to the same folder after a move without having to restart the chunk
	// upload
	p := path.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir,
		i.UploadID())

	if err := os.MkdirAll(p, DirPerm); err != nil {
		return "", err
	}
	return p, nil
}
func (s *local) remove(ctx context.Context, p *storage.RemoveParams) error {
	log := logger.MustFromContext(ctx)

	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	log.Info("local: remove " + ap)

	if p.Recursive == false {
		return s.convertError(os.Remove(ap))
	}
	return s.convertError(os.RemoveAll(ap))
}

func (s *local) getMergedMetaData(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {

	m, err := s.getFSInfo(ctx, p)
	if err != nil {
		return nil, s.convertError(err)
	}

	rec, err := s.aero.GetOrCreateRecord(p.Rsp)
	if err != nil {
		return nil, s.convertError(err)
	}

	m.Modified = uint64(rec.Bins["mtime"].(int))
	m.ETag = rec.Bins["etag"].(string)
	return m, nil

}

func (s *local) stat(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {
	log := logger.MustFromContext(ctx)
	
	log.Debug(fmt.Sprintf("localstrg.stat called with %+v", *p))	

	m, err := s.getMergedMetaData(ctx, p)
	if err != nil {
		return nil, s.convertError(err)
	}

	if !m.IsContainer || p.Children == false {
		return m, nil
	}

	// fns is just the base name
	fns, err := s.getFSChildrenNames(ctx, p)
	if err != nil {
		return nil, s.convertError(err)
	}

	childrenMeta := []*storage.MetaData{}
	for _, fn := range fns {
		pt := path.Join(p.Rsp, path.Clean(fn))

		mergedParams := &storage.StatParams{}
		mergedParams.BaseParams = p.BaseParams
		mergedParams.Rsp = pt

		m, err := s.getMergedMetaData(ctx, mergedParams)
		if err != nil {
			// just log the error
			log.Err(err.Error())
		} else {
			// healthy children are added to the parent
			childrenMeta = append(childrenMeta, m)
		}
	}
	m.Children = childrenMeta
	return m, nil
}

func (s *local) getFSInfo(ctx context.Context, p *storage.StatParams) (*storage.MetaData, error) {
	log := logger.MustFromContext(ctx)
	rp, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	log.Info("local: stat " + ap)

	// Get storage file info.
	finfo, err := os.Stat(ap)
	if err != nil {
		return nil, s.convertError(err)
	}

	id, err := xattr.GetXAttr(ap, XAttrID)
	if err != nil {
		if err == syscall.ENODATA {
			id = []byte(uuid.New())
			err = xattr.SetXAttr(ap, XAttrID, []byte(id), xattr.XAttrCreate)
			if err != nil {
				return nil, err
			}
			err := s.aero.PutRecord(p.Rsp, string(id))
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if len(id) == 0 { // xattr is empty but is set
		id = []byte(uuid.New())
		err = xattr.SetXAttr(ap, XAttrID, []byte(id), xattr.XAttrCreateOrReplace)
		if err != nil {
			return nil, err
		}
		err := s.aero.PutRecord(p.Rsp, string(id))
		if err != nil {
			return nil, err
		}
	}

	mimeType := s.getMimeType(finfo)
	perm := s.getPermissions(finfo)
	parentPath := s.pathWithPrefix(rp)
	if finfo.IsDir() {
		parentPath += "/" // container' path ends with slash
	}

	m := storage.MetaData{
		ID:          string(id),
		Path:        parentPath,
		Size:        uint64(finfo.Size()),
		IsContainer: finfo.IsDir(),
		MimeType:    mimeType,
		Permissions: perm,
	}

	return &m, nil
}

func (s *local) getFSChildrenNames(ctx context.Context, p *storage.StatParams) ([]string, error) {

	log := logger.MustFromContext(ctx)

	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	fd, err := os.Open(ap)
	if err != nil {
		return nil, s.convertError(err)
	}
	defer func() {
		if err := fd.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				ap, err.Error())

			log.Warning(msg)
		}
	}()

	fns, err := fd.Readdirnames(0)
	if err != nil {
		return nil, s.convertError(err)
	}
	return fns, nil
}

func (s *local) createContainer(ctx context.Context, p *storage.CreateContainerParams) error {
	log := logger.MustFromContext(ctx)
	_, ap := s.getRelAndAbsPaths(p.Rsp, p.Idt)

	log.Info("local: createcontainer " + ap)

	err := os.Mkdir(ap, DirPerm)
	if err != nil {
		return s.convertError(err)
	}

	// Set xattrs, on moves they are preserved.
	resourceID := uuid.New()
	err = xattr.SetXAttr(ap, XAttrID, []byte(resourceID), xattr.XAttrCreate)
	if err != nil {
		return err
	}

	return s.aero.PutRecord(p.Rsp, resourceID)
}

func (s *local) createUserHomeDirectory(ctx context.Context, p *storage.CreateUserHomeDirParams) error {
	exists, err := s.isHomeDirCreated(p.Idt)
	if err != nil {
		return s.convertError(err)
	}
	if exists {
		return nil
	}
	homeDir := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir,
		path.Join(p.Idt.IDMID, p.Idt.PID))

	return s.convertError(os.MkdirAll(homeDir, DirPerm))
}
func (s *local) isHomeDirCreated(idt *auth.Identity) (bool, error) {
	homeDir := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir,
		path.Join(idt.IDMID, idt.PID))

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
	return path.Join(s.cfg.GetDirectives().LocalStorageRootTmpDir, uuid.New())
}
func (s *local) commitPutFile(from, to string) error {
	return os.Rename(from, to)
}

func (s *local) stageFile(ctx context.Context, source string, dest string, size int64) (err error) {
	log := logger.MustFromContext(ctx)
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := reader.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				source, err.Error())

			log.Warning(msg)
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

			log.Warning(msg)
		}
	}()

	_, err = io.CopyN(writer, reader, size)
	if err != nil {
		return err
	}
	return nil
}

func (s *local) stageDir(ctx context.Context, source string, dest string) (err error) {
	log := logger.MustFromContext(ctx)
	// create dest dir
	err = os.MkdirAll(dest, DirPerm)
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	defer func() {
		if err := directory.Close(); err != nil {
			msg := fmt.Sprintf("local: cannot close resource:%s err:%s",
				source, err.Error())

			log.Warning(msg)
		}
	}()

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := path.Join(source, obj.Name())
		destinationfilepointer := path.Join(dest, obj.Name())

		if obj.IsDir() {
			// create sub-directories - recursively
			err = s.stageDir(ctx, sourcefilepointer, destinationfilepointer)
			if err != nil {
				return err
			}
		} else {
			// perform copy
			err = s.stageFile(ctx, sourcefilepointer, destinationfilepointer, obj.Size())
			if err != nil {
				return err
			}
		}

	}
	return
}

func (s *local) sanitizePath(rsp string) string {
	return rsp
}

func (s *local) pathWithoutPrefix(rsp string) string {
	parts := strings.Split(rsp, "/")
	if len(parts) == 1 {
		return ""
	} else {
		return strings.Join(parts[1:], "/")
	}
}
func (s *local) pathWithPrefix(rp string) string {
	return path.Join(s.Prefix(), path.Clean(rp))
}

func (s *local) getMimeType(fi os.FileInfo) string {
	if fi.IsDir() {
		return storage.DefaultContainerMimeType
	}
	mimeType := mime.TypeByExtension(path.Ext(fi.Name()))
	if mimeType == "" {
		mimeType = storage.DefaultObjectMimeType
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
func (s *local) getRelAndAbsPaths(rsp string,
	idt *auth.Identity) (string, string) {

	rp := s.pathWithoutPrefix(rsp)
	ap := path.Join(s.cfg.GetDirectives().LocalStorageRootDataDir,
		path.Join(idt.IDMID, idt.PID, rp))

	return rp, ap
}

// storageInfo represents the information obtainable local filesystem.
type storageInfo struct {
	os.FileInfo
	MimeType     string
	Permissions  storage.ResourceMode
	ResourcePath string
}

type chunkPathInfo struct {
	ResourcePath string
	TransferID   string
	TotalChunks  uint64
	CurrentChunk uint64
}

func (c *chunkPathInfo) UploadID() string {
	return "chunking-" + c.TransferID + "-" + strconv.FormatUint(c.TotalChunks, 10)
}

type ChunkHeaderInfo struct {
	// OC-Chunked = 1
	OCChunked bool

	// OC-Chunk-Size
	OCChunkSize uint64

	// OC-Total-Length
	OCTotalLength uint64
}

func (c *chunkPathInfo) String() string {
	return fmt.Sprintf("chunkPathInfo: (%+v)", *c)
}

// IsChunked determines if an upload is chunked or not.
func IsChunked(rsp string) (bool, error) {
	return regexp.MatchString(`-chunking-\w+-[0-9]+-[0-9]+`, rsp)
}

// GetChunkPathInfo obtains the different parts of a chunk from the path.
func GetChunkPathInfo(rsp string) (*chunkPathInfo, error) {
	parts := strings.Split(rsp, "-chunking-")
	tail := strings.Split(parts[1], "-")

	totalChunks, err := strconv.ParseUint(tail[1], 10, 64)
	if err != nil {
		return nil, err
	}
	currentChunk, err := strconv.ParseUint(tail[2], 10, 64)
	if err != nil {
		return nil, err
	}

	if currentChunk >= totalChunks {
		return nil, fmt.Errorf("current chunk:%d exceeds total chunks:%d.", currentChunk, totalChunks)
	}

	info := &chunkPathInfo{}
	info.ResourcePath = parts[0]
	info.TransferID = tail[0]
	info.TotalChunks = totalChunks
	info.CurrentChunk = currentChunk

	return info, nil
}
