package ocwebdav

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	//"github.com/clawio/clawiod/entities"
	"github.com/clawio/clawiod/helpers"
	"github.com/clawio/clawiod/keys"
	"github.com/clawio/clawiod/services/metadata/metadatacontroller/ocsql"

	"github.com/gorilla/mux"
)

func (s *svc) PutChunked(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user := keys.MustGetUser(r)
	log := keys.MustGetLog(r)
	path := mux.Vars(r)["path"]

	chunkInfo, err := getChunkBLOBInfo(path)
	if err != nil {
		log.WithError(err).WithField("pathspec", path).Error("cannot obtain chunk info from pathspec")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.WithField("chunk", chunkInfo).Debug("chunk info")

	chunkTempFilename, chunkTempFile, err := s.createChunkTempFile()
	if err != nil {
		log.WithError(err).Error("cannot create chunk temp file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer chunkTempFile.Close()

	readCloser := http.MaxBytesReader(w, r.Body, int64(s.conf.GetDirectives().WebDAV.UploadMaxFileSize))
	if _, err := io.Copy(chunkTempFile, readCloser); err != nil {
		s.handlePutError(err, w, r)
		return
	}

	// force close of the file here because if it is the last chunk to
	// assemble the big file we need all the chunks closed
	if err = chunkTempFile.Close(); err != nil {
		s.handlePutError(err, w, r)
		return
	}

	chunkFolderName, err := s.getChunkFolderName(chunkInfo)
	if err != nil {
		s.handlePutError(err, w, r)
		return
	}
	log.WithField("chunkfolder", chunkFolderName).Debug("chunk folder info")

	chunkTarget := helpers.SecureJoin(chunkFolderName, fmt.Sprintf("%d", chunkInfo.currentChunk))
	if err = os.Rename(chunkTempFilename, chunkTarget); err != nil {
		s.handlePutError(err, w, r)
		return
	}
	log.WithField("chunktarget", chunkTarget).Debug("chunk target info")

	// Check that all chunks are uploaded.
	// This is very inefficient, the server has to check that it has all the
	// chunks after each uploaded chunk.
	// A two-phase upload like DropBox is better, because the server will
	// assembly the chunks when the client asks for it.
	chunkFolder, err := os.Open(chunkFolderName)
	if err != nil {
		s.handlePutError(err, w, r)
		return
	}
	defer chunkFolder.Close()

	// read all the chunks inside the chunk folder; -1 == all
	chunks, err := chunkFolder.Readdir(-1)
	if err != nil {
		s.handlePutError(err, w, r)
		return
	}
	log.WithField("chunk count", len(chunks)).Debug("current amount of chunks")

	// there is still some chunks to be uploaded so we stop here
	if len(chunks) < int(chunkInfo.totalChunks) {
		log.Debug("chunk is not final")
		w.WriteHeader(http.StatusCreated)
		return
	}

	assembledFileName, assembledFile, err := s.createChunkTempFile()
	if err != nil {
		s.handlePutError(err, w, r)
		return
	}
	defer assembledFile.Close()
	log.WithField("assembledfile", assembledFileName).Debug("assembled file info")

	// walk all chunks and append to assembled file
	for i, _ := range chunks {
		target := helpers.SecureJoin(chunkFolderName, fmt.Sprintf("%d", i))

		chunk, err := os.Open(target)
		if err != nil {
			s.handlePutError(err, w, r)
			return
		}

		if _, err = io.Copy(assembledFile, chunk); err != nil {
			s.handlePutError(err, w, r)
			return
		}
		log.WithField("chunk", target).WithField("assembledfile", assembledFileName).Debug("chunk appended to assembled file")

		// we close the chunk here because if the assemnled file contains hundreds of chunks
		// we will end up with hundreds of open file descriptors
		if err = chunk.Close(); err != nil {
			s.handlePutError(err, w, r)
			return

		}
	}

	// at this point the assembled file is complete
	// so we free space removing the chunks folder
	defer func() {
		if err = os.RemoveAll(chunkFolderName); err != nil {
			log.WithError(err).Error("cannot remove chunk folder after recontruction")
		}
	}()

	// when writing to the assembled file the write pointer points to the end of the file
	// so we need to seek it to the beggining
	if _, err = assembledFile.Seek(0, 0); err != nil {
		s.handlePutError(err, w, r)
		return
	}

	log.WithField("pathspec", chunkInfo.pathSpec).Debug("upload chunk to final destination")
	if err = s.dataController.UploadBLOB(user, chunkInfo.pathSpec, assembledFile, ""); err != nil {
		s.handlePutError(err, w, r)
		return
	}

	info, err := s.metaDataController.ExamineObject(user, chunkInfo.pathSpec)
	if err != nil {
		s.handlePutError(err, w, r)
		return
	}

	w.Header().Add("Content-Type", info.MimeType)
	w.Header().Set("ETag", info.Extra.(ocsql.Extra).ETag)
	w.Header().Set("OC-FileId", info.Extra.(ocsql.Extra).ID)
	w.Header().Set("OC-ETag", info.Extra.(ocsql.Extra).ETag)
	t := time.Unix(info.ModTime/1000000000, info.ModTime%1000000000)
	lastModifiedString := t.Format(time.RFC1123)
	w.Header().Set("Last-Modified", lastModifiedString)
	w.Header().Set("X-OC-MTime", "accepted")

	// if object did not exist, http code is 201, else 204.
	if info == nil {
		w.WriteHeader(http.StatusCreated)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type chunkHeaderInfo struct {
	// OC-Chunked = 1
	ochunked bool

	// OC-Chunk-Size
	ocChunkSize uint64

	// OC-Total-Length
	ocTotalLength uint64
}

type chunkBLOBInfo struct {
	pathSpec     string
	transferID   string
	totalChunks  int64
	currentChunk int64
}

// not using the resource path in the chunk folder name allows uploading
// to the same folder after a move without having to restart the chunk
// upload
func (c *chunkBLOBInfo) uploadID() string {
	return fmt.Sprintf("chunking-%s-%d", c.transferID, c.totalChunks)
}

func getChunkBLOBInfo(pathSpec string) (*chunkBLOBInfo, error) {
	parts := strings.Split(pathSpec, "-chunking-")
	tail := strings.Split(parts[1], "-")

	totalChunks, err := strconv.ParseInt(tail[1], 10, 64)
	if err != nil {
		return nil, err
	}

	currentChunk, err := strconv.ParseInt(tail[2], 10, 64)
	if err != nil {
		return nil, err
	}

	if currentChunk >= totalChunks {
		return nil, fmt.Errorf("current chunk:%d exceeds total number of chunks:%d.", currentChunk, totalChunks)
	}

	return &chunkBLOBInfo{
		pathSpec:     parts[0],
		transferID:   tail[0],
		totalChunks:  totalChunks,
		currentChunk: currentChunk,
	}, nil
}

func (s *svc) createChunkTempFile() (string, *os.File, error) {
	dirs := s.conf.GetDirectives()

	file, err := ioutil.TempFile(dirs.OCWebDAV.ChunksTemporaryNamespace, "")
	if err != nil {
		return "", nil, err
	}

	fn := helpers.SecureJoin(file.Name())
	return fn, file, nil
}

func (s *svc) getChunkFolderName(i *chunkBLOBInfo) (string, error) {
	dirs := s.conf.GetDirectives()
	p := helpers.SecureJoin(dirs.OCWebDAV.ChunksTemporaryNamespace, i.uploadID())
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}
