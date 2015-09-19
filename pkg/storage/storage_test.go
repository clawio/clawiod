// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage_test

import (
	"github.com/clawio/clawiod/pkg/config"
	. "gopkg.in/check.v1"
	"testing"

	// Storages to test
	localstorage "github.com/clawio/clawiod/pkg/storage/providers/local"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type StorageSuite struct{}

var _ = Suite(&StorageSuite{})

func (s *StorageSuite) SetUpSuite(c *C) {
	local, err := localstorage.New("local", cfg, log)
}

func (s *StorageSuite) TestGetStorage(c *C) {

}
