// Copyright © 2015 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

var countFlag int
var bsFlag int
var checksumFlag string
var cernDistributionFlag bool
var randomTargetFlag bool

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Benchmarks the uploading process using different file sizes",
	RunE:  upload,
	Long: `This benchmark test will measure the upload performance.

The object size is the result of block size x count. This is the same
approach used by dd.`,
}

// createFile is a substitute for dd
// char is the character to insert
// count is the number of blocks
// bs is the block size: how many bytes are we going to write flush every round.
func createFile(fn, char string, count, bs int) (*os.File, error) {
	logger := getLogger()
	var fd *os.File
	if fn == "" {
		tf, err := ioutil.TempFile("", "CLAWIOBENCH-")
		if err != nil {
			return nil, err
		}
		fd = tf
	} else {
		tf, err := os.Create(path.Join(os.TempDir(), fn))
		if err != nil {
			logger.Error().Log("error", err)
			return nil, err
		}
		fd = tf
	}

	// if char is 1 byte then the buffer size will be equal to bs
	buffer := bytes.Repeat([]byte(char), bs)

	for i := 0; i < count; i++ {
		_, err := fd.Write(buffer)
		if err != nil {
			return nil, err
		}
	}

	return fd, nil
}

// This is the distribution of files at CERN
func createCERNDistribution() ([]string, error) {
	fds := []*os.File{}
	fns := []string{}
	fd50MB, err := createFile("testfile-50MB", "1", 1024, 1024*50)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd50MB)

	fd15MB, err := createFile("testfile-15MB", "1", 1024, 1024*15)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd15MB)

	fd8MB, err := createFile("testfile-8MB", "1", 1024, 1024*8)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd8MB)

	fd10MB, err := createFile("testfile-8MB", "1", 1024, 1024*10)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd10MB)

	fd5MB, err := createFile("testfile-5MB", "1", 1024, 1024*5)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd5MB)

	fd4MB, err := createFile("testfile-4MB", "1", 1024, 1024*4)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd4MB)

	fd3MB, err := createFile("testfile-3MB", "1", 1024, 1024*3)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd3MB)

	fd2MB, err := createFile("testfile-2MB", "1", 1024, 1024*2)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd2MB)

	fd1MB, err := createFile("testfile-1MB", "1", 1024, 1024)
	if err != nil {
		return fns, err
	}
	fds = append(fds, fd1MB)

	for i := 0; i < 11; i++ {
		fn := fmt.Sprintf("testfile-500KB-%d", i)
		fd, err := createFile(fn, "1", 1024, 500)
		if err != nil {
			return fns, err
		}
		fds = append(fds, fd)
	}

	for i := 0; i < 32; i++ {
		fn := fmt.Sprintf("testfile-50KB-%d", i)
		fd, err := createFile(fn, "1", 1024, 50)
		if err != nil {
			return fns, err
		}
		fds = append(fds, fd)
	}

	for i := 0; i < 28; i++ {
		fn := fmt.Sprintf("testfile-5KB-%d", i)
		fd, err := createFile(fn, "1", 1024, 5)
		if err != nil {
			return fns, err
		}
		fds = append(fds, fd)
	}

	for i := 0; i < 15; i++ {
		fn := fmt.Sprintf("testfile-1KB-%d", i)
		fd, err := createFile(fn, "1", 1024, 1)
		if err != nil {
			return fns, err
		}
		fds = append(fds, fd)
	}

	for i := 0; i < 5; i++ {
		fn := fmt.Sprintf("testfile-100B-%d", i)
		fd, err := createFile(fn, "1", 1, 100)
		if err != nil {
			return fns, err
		}
		fds = append(fds, fd)
	}

	for _, v := range fds {
		fns = append(fns, v.Name())
		v.Close()
	}

	return fns, nil
}

func upload(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		cmd.Help()
		return nil
	}

	logger := getLogger()

	if concurrencyFlag > probesFlag {
		concurrencyFlag = probesFlag
	}
	if concurrencyFlag == 0 {
		concurrencyFlag++
	}

	var fns []string
	if cernDistributionFlag {
		vals, err := createCERNDistribution()
		if err != nil {
			return err
		}
		fns = vals
	} else {
		fd, err := createFile(fmt.Sprintf("testfile-manual-count-%d-bs-%d", countFlag, bsFlag), "1", countFlag, bsFlag)
		if err != nil {
			return err
		}
		fns = []string{fd.Name()}
		fd.Close()
	}

	defer func() {
		for _, v := range fns {
			os.RemoveAll(v)
		}
	}()

	benchStart := time.Now()

	total := 0
	errorProbes := 0

	errChan := make(chan error)
	resChan := make(chan string)
	doneChan := make(chan bool)
	limitChan := make(chan int, concurrencyFlag)

	for i := 0; i < concurrencyFlag; i++ {
		limitChan <- 1
	}

	var bar *pb.ProgressBar
	if progressBar {
		fmt.Printf("There are %d possible files to upload\n", len(fns))
		bar = pb.StartNew(probesFlag)
	}

	for i := 0; i < probesFlag; i++ {
		rand.Seed(time.Now().UnixNano())
		filename := fns[rand.Intn(len(fns))]
		go func(fn string) {
			<-limitChan
			defer func() {
				limitChan <- 1
			}()

			// open again the file
			lfd, err := os.Open(fn)
			if err != nil {
				errChan <- err
				return
			}
			defer lfd.Close()

			c := &http.Client{} // connections are reused if we reuse the client
			// PUT will close the fd
			// is it possible that the HTTP client is reusing connections so is being blocked?
			target := args[0]
			if randomTargetFlag {
				target += uuid.NewV4().String()
			}
			req, err := http.NewRequest("PUT", os.Getenv("CLAWIOBENCH_OCWEBDAV")+"/"+target, lfd)
			if err != nil {
				errChan <- err
				return
			}

			req.Header.Add("Content-Type", "application/octet-stream")
			req.SetBasicAuth(os.Getenv("CLAWIOBENCH_USERNAME"), os.Getenv("CLAWIOBENCH_PASSWORD"))

			res, err := c.Do(req)
			if err != nil {
				errChan <- err
				return
			}

			err = res.Body.Close()
			if err != nil {
				errChan <- err
				return
			}

			if res.StatusCode ==  http.StatusCreated || res.StatusCode == http.StatusNoContent {
				doneChan <- true
				resChan <- ""
				return
			}
			err = fmt.Errorf("Request failed with status code %d", res.StatusCode)
			errChan <- err
			return
		}(filename)
	}

	for {
		select {
		case _ = <-doneChan:
			total++
			if progressBar {
				bar.Increment()
			}
		case _ = <-resChan:
		case err := <-errChan:
			logger.Error().Log("error", err)
			errorProbes++
			total++
			if progressBar {
				bar.Increment()
			}
		}

		if total == probesFlag {
			break
		}
	}

	if progressBar {
		bar.Finish()
	}

	numberRequests := probesFlag
	concurrency := concurrencyFlag
	totalTime := time.Since(benchStart).Seconds()
	failedRequests := errorProbes
	frequency := float64(numberRequests-failedRequests) / totalTime
	period := float64(1 / frequency)
	volume := numberRequests * countFlag * bsFlag / 1024 / 1024
	throughput := float64(volume) / totalTime
	data := [][]string{
		{"#NUMBER", "CONCURRENCY", "TIME", "FAILED", "FREQ", "PERIOD", "VOLUME", "THROUGHPUT"},
		{fmt.Sprintf("%d", numberRequests), fmt.Sprintf("%d", concurrency), fmt.Sprintf("%f", totalTime), fmt.Sprintf("%d", failedRequests), fmt.Sprintf("%f", frequency), fmt.Sprintf("%f", period), fmt.Sprintf("%d", volume), fmt.Sprintf("%f", throughput)},
	}
	w := csv.NewWriter(output)
	w.Comma = ' '
	for _, d := range data {
		if err := w.Write(d); err != nil {
			return err
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return err
	}
	return nil
}

func init() {
	RootCmd.AddCommand(uploadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uploadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uploadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	uploadCmd.Flags().IntVar(&countFlag, "count", 1024, "The number of blocks of the file")
	uploadCmd.Flags().IntVar(&bsFlag, "bs", 1024, "The number of bytes of each block")
	uploadCmd.Flags().StringVar(&checksumFlag, "checksum", "", "The checksum for the file")
	uploadCmd.Flags().BoolVar(&cernDistributionFlag, "cern-distribution", false, "Use file sizes that follow the distribution found on CERNBox")
	uploadCmd.Flags().BoolVar(&randomTargetFlag, "random-target", false, "Add a random value to the upload target filename")

}
