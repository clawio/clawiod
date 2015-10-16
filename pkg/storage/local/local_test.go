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

type LocalSuite struct{}

var _ = Suite(&LocalSuite{})

var tests = map[string]*chunkPathInfo{
	"test-chunking-aljsd938429-32-31": &chunkPathInfo{
		ResourcePath: "test",
		TotalChunks:  32,
		CurrentChunk: 31,
		TransferID:   "aljsd938429",
	},
	"/docs/thesis.pdf-chunking-__abc123__-10-0": &chunkPathInfo{
		ResourcePath: "/docs/thesis.pdf",
		TotalChunks:  10,
		CurrentChunk: 0,
		TransferID:   "__abc123__",
	},
	"local/test-chunking-tid-2-0": &chunkPathInfo{
		ResourcePath: "local/test",
		TotalChunks:  2,
		CurrentChunk: 0,
		TransferID:   "tid",
	},
}

func (s *LocalSuite) TestIsChunked(c *C) {
	for p := range tests {
		ok, err := IsChunked(p)
		if err != nil || !ok {
			c.Error("it should be chunked:", err)
			return
		}
	}
}

func (s *LocalSuite) TestGetChunkPathInfo(c *C) {
	for p, i := range tests {
		info, err := GetChunkPathInfo(p)
		if err != nil {
			c.Error(err)
			return
		}
		c.Assert(info.ResourcePath, Equals, i.ResourcePath)
		c.Assert(info.TotalChunks, Equals, i.TotalChunks)
		c.Assert(info.CurrentChunk, Equals, i.CurrentChunk)
		c.Assert(info.TransferID, Equals, i.TransferID)
	}
}
