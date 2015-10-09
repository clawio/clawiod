package xattr

import (
	"github.com/clawio/clawiod/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"path"
	"syscall"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type XAttrSuite struct{ tmpDir string }

var _ = Suite(&XAttrSuite{})

func (s *XAttrSuite) createFile() string {
	fn := path.Join(s.tmpDir, uuid.New())
	ioutil.WriteFile(fn, []byte(`some data`), 0644)
	return fn
}

func (s *XAttrSuite) SetUpSuite(c *C) {
	s.tmpDir = c.MkDir()
}

func (s *XAttrSuite) TestGetXAttr(c *C) {
	fn := s.createFile()
	val := []byte("clawioid1234567890")
	SetXAttr(fn, "user.cid", val, XAttrCreate)

	attrVal, err := GetXAttr(fn, "user.cid")
	if err != nil {
		c.Error(err)
	}

	c.Assert(string(attrVal), Equals, string(val))
}

func (s *XAttrSuite) TestSetXAtttr(c *C) {
	fn := s.createFile()
	err := SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrCreate)
	if err != nil {
		c.Error(err)
		return
	}
}

func (s *XAttrSuite) TestSetXAttrUpsert(c *C) {
	fn := s.createFile()
	err := SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrCreate)
	if err != nil {
		c.Error(err)
		return
	}

	if err := SetXAttr(fn, "user.cid", []byte("other"),
		XAttrCreateOrReplace); err != nil {
		c.Error(err)
	}
}

func (s *XAttrSuite) TestSetXAttrReplace(c *C) {
	fn := s.createFile()
	err := SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrCreate)
	if err != nil {
		c.Error(err)
		return
	}

	if err := SetXAttr(fn, "user.cid", []byte("other"),
		XAttrReplace); err != nil {
		c.Error(err)
	}
}

// FAILURE SCENARIOS
func (s XAttrSuite) TestGetXAttrFailNoPath(c *C) {
	_, err := GetXAttr("nopath", "user.notexists")
	if err != syscall.ENOENT {
		c.Error(err)
	}
}

func (s XAttrSuite) TestGetXAttrFail(c *C) {
	fn := s.createFile()
	_, err := GetXAttr(fn, "user.notexists")
	if err != syscall.ENODATA {
		c.Error(err)
	}
}

func (s *XAttrSuite) TestSetXAtttrFail(c *C) {
	fn := s.createFile()
	err := SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrCreate)
	if err != nil {
		c.Error(err)
		return
	}

	// MUST fail with EEXIST error
	err = SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrCreate)
	if err != syscall.EEXIST {
		c.Error(err)
	}
}

func (s *XAttrSuite) TestSetXAtttrReplaceFail(c *C) {
	fn := s.createFile()
	err := SetXAttr(fn, "user.cid", []byte("clawioid1234567890"), XAttrReplace)
	if err != syscall.ENODATA {
		c.Error(err)
	}
}
