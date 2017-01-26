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
	"os"

	"github.com/spf13/cobra"
	"io"
)

var probesFlag int
var concurrencyFlag int
var csvFile string
var progressBar bool
var logfile string

var output io.Writer

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "clawiowebdavbench",
	Short: "ClawIO WebDAV Benchmarking Tool",
	Long: `clawiowebdavbench is a tool for benchmarking WebDAV endpoints.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().IntVarP(&probesFlag, "requests", "n", 1, "Number of requests to perform for the benchmarking session. The default is to just perform a single request which usually leads to non-representative benchmarking results.")
	RootCmd.PersistentFlags().IntVarP(&concurrencyFlag, "concurrency", "c", 1, "Number of multiple requests to perform at a time. Default is one request at a time.")
	RootCmd.PersistentFlags().StringVarP(&csvFile, "csv-file", "", "", "Write the results to  a Comma separated value (CSV) file.")
	RootCmd.PersistentFlags().BoolVar(&progressBar, "progress-bar", true, "Show progress bar")
	RootCmd.PersistentFlags().StringVarP(&logfile, "log-file", "", "", "Write errors to this file. It logs to stder by default")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if csvFile != "" {
		fd, err := os.Create(csvFile)
		if err != nil {
			fmt.Printf("Cannot open csv file: %s\n", err.Error())
			os.Exit(1)
		}
		output = fd
	} else {
		output = os.Stdout
	}
}
