package ocfsdatadriver

import (
	"context"
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
	"strings"

	"github.com/clawio/clawiod/root"
	"github.com/clawio/clawiod/root/ocfsmdatadriver"
	"github.com/go-kit/kit/log/levels"
)

type driver struct {
	logger                 levels.Levels
	dataFolder             string
	temporaryFolder        string
	checksum               string
	verifyClientChecksum   bool
	metaDataDriver         root.MetaDataDriver
	ownCloudMetaDataDriver *ocfsmdatadriver.Driver
}

// New returns an implementation of DataDriver.
func New(logger levels.Levels, dataFolder, temporaryFolder, checksum string, verifyClientChecksum bool, metaDataDriver root.MetaDataDriver) (root.DataDriver, error) {
	if err := os.MkdirAll(dataFolder, 755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(temporaryFolder, 0755); err != nil {
		return nil, err
	}

	// check that metadata driver is compatible with owncloud.
	// This is an ugly check but is the one needed to keep the metadata driver interface clean
	// We cast to all compatible implementations.
	ownCloudMetaDataDriver, ok := metaDataDriver.(*ocfsmdatadriver.Driver)
	if !ok {
		logger.Crit().Log("error", "metadata driver is not ocfsmdatadriver")
		return nil, errors.New("metadata driver is not ocfsmdatadriver")
	}

	return &driver{
		logger:                 logger,
		dataFolder:             strings.Trim(dataFolder, "/"),
		temporaryFolder:        strings.Trim(temporaryFolder, "/"),
		checksum:               checksum,
		verifyClientChecksum:   verifyClientChecksum,
		metaDataDriver:         metaDataDriver,
		ownCloudMetaDataDriver: ownCloudMetaDataDriver,
	}, nil
}

func (c *driver) Init(ctx context.Context, user root.User) error {
	return nil
}

// UploadFile saves a file to disk.
// This operation has 4 phases:
// 1) Write the file to a temporary folder.
// 2) Optional: calculate the checksum of the file if server-checksum is enabled.
// 3) Optional: if a client-checksum is provided, check if it matches with the server-checksum.
// 4) Move the file from the temporary folder to user folder.
func (c *driver) UploadFile(ctx context.Context, user root.User, path string, r io.ReadCloser, clientChecksum string) error {
	tempFileName, err := c.saveToTempFile(r)
	if err != nil {
		c.logger.Error().Log("error", err)
		return err
	}
	defer r.Close()

	var computedChecksum string
	// 2) Optional: calculate the checksum of the file.
	if c.checksum != "" {
		computedChecksum, err = c.computeChecksum(tempFileName)
		if err != nil {
			c.logger.Error().Log("error", err)
			return err
		}

		c.logger.Info().Log("msg", "checksum computed", "checksum", computedChecksum, "file", tempFileName)

		// 3) Optional: verify if server-checksum matches client-checksum.
		if c.verifyClientChecksum {
			if computedChecksum != clientChecksum {
				msg := fmt.Sprintf("fsdatadriver: wrong checksum computed:%q received:%q",
					computedChecksum, clientChecksum)
				return checksumError(msg)
			}
		}
	}

	// 4) Move the file from the temporary folder to user folder.
	localPath := c.getLocalPath(user, path)
	if err := os.Rename(tempFileName, localPath); err != nil {
		c.logger.Error().Log("error", err)
		if os.IsNotExist(err) {
			return notFoundError(err.Error())
		}
		return err
	}
	c.logger.Info().Log("msg", "atomic rename completed", "source", tempFileName, "target", localPath)
	if err = c.ownCloudMetaDataDriver.PropagateChanges(user, path, "/", computedChecksum); err != nil {
		c.logger.Error().Log("error", err, "msg", "error propagating changes")
	}
	return nil
}

func (c *driver) DownloadFile(ctx context.Context, user root.User, path string) (io.ReadCloser, error) {
	localPath := c.getLocalPath(user, path)
	fd, err := os.Open(localPath)
	if err != nil {
		c.logger.Error().Log("error", err)
		if os.IsNotExist(err) {
			return nil, notFoundError(err.Error())
		}
		return nil, err
	}
	fileInfo, err := fd.Stat()
	if err != nil {
		c.logger.Error().Log("error", err)
		if os.IsNotExist(err) {
			return nil, notFoundError(err.Error())
		}
		return nil, err
	}

	if fileInfo.IsDir() {
		err := isFolderError("file is a folder")
		c.logger.Error().Log("error", err)
		return nil, err
	}

	c.logger.Info().Log("msg", "file opened for reading", "file", localPath)
	return fd, nil
}

func (c *driver) saveToTempFile(r io.Reader) (string, error) {
	temporaryFolder := fmt.Sprintf("/%s", c.temporaryFolder)
	fd, err := ioutil.TempFile(temporaryFolder, "")
	defer fd.Close()
	if err != nil {
		return "", err
	}

	written, err := io.Copy(fd, r)
	if err != nil {
		return "", err
	}

	c.logger.Error().Log("msg", "file written to temporary file", "wb", written, "file", fd.Name())
	return fd.Name(), nil
}

func (c *driver) computeChecksum(fn string) (string, error) {
	checksumType := strings.ToLower(c.checksum)
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
		return "", errors.New(fmt.Sprintf("fsdatadriver: provided checksum %q not implemented", c.checksum))
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

func (c *driver) getLocalPath(user root.User, path string) string {
	path = strings.Trim(path, "/")
	return fmt.Sprintf("/%s/%s/%s/%s", c.dataFolder, string(user.Username()[0]), user.Username(), path)
}

type checksumError string

func (e checksumError) Error() string {
	return string(e)
}
func (e checksumError) Code() root.Code {
	return root.Code(root.CodeBadChecksum)
}
func (e checksumError) Message() string {
	return string(e)
}

type notFoundError string

func (e notFoundError) Error() string {
	return string(e)
}
func (e notFoundError) Code() root.Code {
	return root.Code(root.CodeNotFound)
}
func (e notFoundError) Message() string {
	return string(e)
}

type isFolderError string

func (e isFolderError) Error() string {
	return string(e)
}
func (e isFolderError) Code() root.Code {
	return root.Code(root.CodeBadInputData)
}
func (e isFolderError) Message() string {
	return string(e)
}
