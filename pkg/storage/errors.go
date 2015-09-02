// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

// ExistError represents the error ocurred when the resource already exists.
type ExistError struct {
	Err string
}

func (e *ExistError) Error() string { return e.Err }

// NotExistError represents the error ocurred when the resource does not exist.
type NotExistError struct {
	Err string
}

func (e *NotExistError) Error() string { return e.Err }

// UnsupportedChecksumTypeError represents the error when trying to put a file with a checksum that is not supported
type UnsupportedChecksumTypeError struct {
	Err string
}

func (e *UnsupportedChecksumTypeError) Error() string {
	return e.Err
}

// BadChecksumError represents the error that happens when client and server checksums missmatch.
type BadChecksumError struct {
	Err string
}

func (e *BadChecksumError) Error() string {
	return e.Err
}

// ThirdPartyCopyNotEnabled represents the error that happen when trying
// to do a third party copy (src and dst have different storage schemes) and it is not enabled
// in the storage.
// Third party copy permissions is in and out.
type ThirdPartyCopyNotEnabled struct {
	Err string
}

func (e *ThirdPartyCopyNotEnabled) Error() string { return "third party copy not enabled" }

// ThirdPartyRenameNotEnabled represents the error that happen when trying
// to do a third party copy (src and dst have different storage schemes) and it is not enabled
// in the storage.
// Third party copy permissions is in and out.
type ThirdPartyRenameNotEnabled struct {
	Err string
}

func (e *ThirdPartyRenameNotEnabled) Error() string { return "third party rename not enabled" }
