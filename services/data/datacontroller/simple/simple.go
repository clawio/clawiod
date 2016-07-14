package simple

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"hash/adler32"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/services/data/datacontroller"

	"github.com/Sirupsen/logrus"
)

type simpleDataController struct {
	conf *config.Config
	log  *logrus.Entry
}

// New returns an implementation of DataController.
func New(conf *config.Config) (datacontroller.DataController, error) {
	dirs := conf.GetDirectives()
	// create namespace and temporary namespace
	if err := os.MkdirAll(dirs.Data.Simple.Namespace, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dirs.Data.Simple.TemporaryNamespace, 0755); err != nil {
		return nil, err
	}
	return &simpleDataController{
		conf: conf,
		log:  helpers.GetAppLogger(conf).WithField("module", "simple:datacontroller"),
	}, nil
}

// UploadBLOB saves a blob to disk.
// This operation has 4 phases:
// 1) Write the blob to a temporary directory.
// 2) Optional: calculate the checksum of the blob if server-checksum is enabled.
// 3) Optional: if a client-checksum is provided, check if it matches with the server-checksum.
// 4) Move the blob from the temporary directory to user directory.
func (c *simpleDataController) UploadBLOB(user *entities.User, pathSpec string, r io.Reader, clientchecksum string) error {
	tempFileName, err := c.saveToTempFile(r)
	if err != nil {
		return err
	}

	// 2) Optional: calculate the checksum of the blob.
	if c.conf.GetDirectives().Data.Simple.Checksum != "" {
		serverchecksum, err := c.computeChecksum(tempFileName)
		if err != nil {
			return err
		}

		// 3) Optional: verify if server-checksum matches client-checksum.
		if c.conf.GetDirectives().Data.Simple.VerifyClientChecksum {
			if serverchecksum != clientchecksum {
				msg := fmt.Sprintf("checksums differ. serverchksum:%q clientchksum:%q",
					serverchecksum, clientchecksum)
				return codes.NewErr(codes.BadChecksum, msg)
			}
		}
	}

	// 4) Move the blob from the temporary directory to user directory.
	storagePath := c.getStoragePath(user, pathSpec)
	if err := os.Rename(tempFileName, storagePath); err != nil {
		if os.IsNotExist(err) {
			return codes.NewErr(codes.NotFound, err.Error())
		}
		return err
	}
	return nil
}

func (c *simpleDataController) DownloadBLOB(user *entities.User, pathSpec string) (io.Reader, error) {
	storagePath := c.getStoragePath(user, pathSpec)
	c.log.WithField("storagepath", storagePath).Debug("object to be downloaded")
	fd, err := os.Open(storagePath)
	if err != nil {
		c.log.Error(err)
		if os.IsNotExist(err) {
			return nil, codes.NewErr(codes.NotFound, err.Error())
		}
		return nil, err
	}
	info, err := fd.Stat()
	if err != nil {
		c.log.Error(err)
		return nil, err
	}

	if info.IsDir() {
		err := codes.NewErr(codes.BadInputData, "object is not a blob")
		c.log.Error(err)
		return nil, err
	}
	return fd, nil
}

func (c *simpleDataController) saveToTempFile(r io.Reader) (string, error) {
	fd, err := ioutil.TempFile(c.conf.GetDirectives().Data.Simple.TemporaryNamespace, "")
	defer fd.Close()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fd, r); err != nil {
		return "", err
	}
	return fd.Name(), nil
}

func (c *simpleDataController) computeChecksum(fn string) (string, error) {
	checksumType := strings.ToLower(c.conf.GetDirectives().Data.Simple.Checksum)
	var hash hash.Hash
	switch checksumType {
	case "md5":
		hash = md5.New()
	case "adler32":
		hash = adler32.New()
	case "sha1":
		hash = sha1.New()
	case "sha256":
		hash = sha256.New()
	default:
		return "", errors.New("provided checksum not implemented")
	}
	fd, err := os.Open(fn)
	defer fd.Close()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(hash, fd); err != nil {
		return "", err
	}
	checksum := fmt.Sprintf("%x", hash.Sum([]byte{}))
	return checksumType + ":" + checksum, nil
}

func (c *simpleDataController) getStoragePath(user *entities.User, path string) string {
	homeDir := secureJoin("/", string(user.Username[0]), user.Username)
	userPath := secureJoin(homeDir, path)
	return secureJoin(c.conf.GetDirectives().Data.Simple.Namespace, userPath)
}

// secureJoin avoids path traversal attacks when joinning paths.
func secureJoin(args ...string) string {
	if len(args) > 1 {
		s := []string{"/"}
		s = append(s, args[1:]...)
		jailedPath := filepath.Join(s...)
		return filepath.Join(args[0], jailedPath)
	}
	return filepath.Join(args...)
}
