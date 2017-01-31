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
	"encoding/csv"
	"fmt"
	br "github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
	"io/ioutil"
)

var childrenFlag bool

var statCmd = &cobra.Command{
	Use:   "propfind <path>",
	Short: "Benchmark getting resource information using PROPFIND",
	RunE:  stat,
}

func stat(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		cmd.Help()
		return nil
	}

	logger := getLogger()

        tr := &http.Transport{
		MaxIdleConnsPerHost: concurrencyFlag,
        }
	client := &http.Client{
		Transport: tr,
	}

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

	var bar *br.ProgressBar
	if progressBar {
		bar = br.StartNew(probesFlag)
	}

	for i := 0; i < probesFlag; i++ {
		go func() {
			<-limitChan
			defer func() {
				limitChan <- 1
			}()

			req, err := http.NewRequest("PROPFIND", os.Getenv("CLAWIOBENCH_OCWEBDAV")+"/", nil)
			if err != nil {
				logger.Error().Log("error", err)
				os.Exit(1)
			}
			req.Close = true

			req.SetBasicAuth(os.Getenv("CLAWIOBENCH_USERNAME"), os.Getenv("CLAWIOBENCH_PASSWORD"))
			req.Header.Set("Depth", "1")
			res, err := client.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer res.Body.Close()
			ioutil.ReadAll(res.Body)

			if res.StatusCode == http.StatusNotFound {
				logger.Error().Log("error", err)
				errChan <- err
				return
			}

			if res.StatusCode == 207 {
				doneChan <- true
				resChan <- ""
				return
			}

			errChan <- fmt.Errorf("internal error: status=%d", res.StatusCode)
		}()
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

	data := [][]string{
		{"number-requests", "concurrency", "time", "failed-requests", "frequency", "period"},
		{fmt.Sprintf("%d", numberRequests), fmt.Sprintf("%d", concurrency), fmt.Sprintf("%f", totalTime), fmt.Sprintf("%d", failedRequests), fmt.Sprintf("%f", frequency), fmt.Sprintf("%f", period)},
	}
	w := csv.NewWriter(output)
	w.Comma = ','
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
	RootCmd.AddCommand(statCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	statCmd.Flags().BoolVarP(&childrenFlag, "children", "", false, "Show children objects inside container")
}
