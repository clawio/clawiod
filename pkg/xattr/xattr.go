// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

// Package xattr handles the manipulation of Linux extended attributes.
// On most Linux systems the user xattrs has the prefix "user".
package xattr

import (
	"syscall"
)

const (
	// XAttrValueBufferSize specifies
	// the maximum length an xattr is truncated when getting it.
	XAttrValueBufferSize = 255
)
const (

	// XAttrCreateOrReplace creates or update the xattr
	XAttrCreateOrReplace = iota

	// XAttrCreate is used to fail when the xattr already exists.
	XAttrCreate

	// XAttrReplace is used to fail when the xattr does not exist.
	XAttrReplace
)

// GetXAttr returns the extended attribute value from the path.
func GetXAttr(path, name string) ([]byte, error) {
	var attr = make([]byte, XAttrValueBufferSize)
	size, err := syscall.Getxattr(path, name, attr)
	if err != nil {
		return attr, err
	}
	return attr[0:size], nil
}

// SetXAttr add an extended attribute to path with a write polciy specified
// in the flags parameter.
// The flags argument can be used to refine the semantics of the operation.
// XATTR_CREATE specifies a pure create, which fails if the named attribute
// exists already.
// XATTR_REPLACE specifies a pure replace operation, which fails if the named
// attribute does not already exist.
// By default (no flags), the extended attribute will be created if need be,
// or will simply replace the value if the attribute exists.
func SetXAttr(path, name string, val []byte, flags int) error {

	err := syscall.Setxattr(path, name, val, flags)
	if err != nil {
		return err
	}
	return nil
}
