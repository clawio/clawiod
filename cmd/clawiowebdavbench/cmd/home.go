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
/*


import (
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Create user home directory",
	Run:   home,
}

func home(cmd *cobra.Command, args []string) {
	logger := getLogger()
	req, err := http.NewRequest("MKCOL", os.Getenv("CLAWIOBENCH_OCWEBDAV")+"/", nil)
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	client := http.DefaultClient
	req.SetBasicAuth(os.Getenv("CLAWIOBENCH_USERNAME"), os.Getenv("CLAWIOBENCH_PASSWORD"))
	res, err := client.Do(req)
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusCreated {
		os.Exit(0)
	}

	if res.StatusCode == http.StatusNotFound {
		logger.Error().Log("error", "not found")
		os.Exit(1)
	}

	logger.Error().Log("error", "internal error: http status code=%d", res.StatusCode)
	os.Exit(1)
}

func init() {
	RootCmd.AddCommand(homeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// homeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// homeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
*/
