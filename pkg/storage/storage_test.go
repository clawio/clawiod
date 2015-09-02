// ClawIO - Scalable Distributed High-Performance Synchronisation and Sharing Service
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package storage

/*
import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

var GetURIFromPathTests = []struct {
	path     string
	expected string
}{
	{"http:/example.org", "http"},
	{"local://abc:abc@some/data", "local"},
	{"swift://abc:abc@example.org/photos", "swift:"},
}

func (s *TestSuite) TestGetURIFromPath(c *C) {
	for _, t := range GetURIFromPathTests {
		uri, err := GetURIFromPath(t.path)
		if err != nil {
			c.Error(err)
		}
		c.Assert(uri.Scheme, Equals, t.expected)
	}
}
*/
