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
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print env vars used by the cli",
	Run:   env,
}

func env(cmd *cobra.Command, args []string) {
	fmt.Printf("export CLAWIOBENCH_OCWEBDAV=%s\n", os.Getenv("CLAWIOBENCH_OCWEBDAV"))
	fmt.Printf("export CLAWIOBENCH_USERNAME=%s\n", os.Getenv("CLAWIOBENCH_USERNAME"))
	fmt.Printf("export CLAWIOBENCH_PASSWORD=%s\n", os.Getenv("CLAWIOBENCH_PASSWORD"))
}

func init() {
	RootCmd.AddCommand(envCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// envCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// envCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
