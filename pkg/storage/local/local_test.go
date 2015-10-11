// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package local

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type LocalSuite struct {}

var _ = Suite(&LocalSuite{})

var fns = []string{
	"test-chunking-aljsd938429-32-32",
	 "/docs/thesis.pdf-chunking-__abc123__-10-0",
}	

func (s *LocalSuite) TestIsChunked(c *C) {
	for _, p := range fns {
		ok, err := isChunked(p)
		if err != nil || !ok {
			c.Error("is should be chunked:", err)
			return
		}
	}
}

func (s *LocalSuite) TestGetChunkPathInfo(c *C) {
	for _, p := range fns {
		info, err := getChunkPathInfo(p)
		if err != nil {
			c.Error(err)
			return
		}
		c.Log(info)
	}
}
