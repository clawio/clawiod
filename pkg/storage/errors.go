// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

// AlreadyExistError represents the error ocurred when the resource already exists.
type AlreadyExistError struct {
	Err string
}

func (e *AlreadyExistError) Error() string { return e.Err }

// NotExistError represents the error ocurred when the resource does not exist.
type NotExistError struct {
	Err string
}

func (e *NotExistError) Error() string { return e.Err }

// BadChecksumError represents the error of a missmatch between client and server checksums.
type BadChecksumError struct {
	Err string
}

func (e *BadChecksumError) Error() string {
	return e.Err
}
