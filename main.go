package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	"github.com/clawio/clawiod/config/file"
	"github.com/clawio/clawiod/daemon"

	"github.com/Sirupsen/logrus"
)

const appName = "clawiod"

var log = logrus.WithField("module", "main")

// Flags that control program flow or startup
var (
	conf    string
	version bool
)

// Build information obtained with the help of -ldflags
var (
	buildDate     string // date -u
	gitTag        string // git describe --exact-match HEAD
	gitNearestTag string // git describe --abbrev=0 --tags HEAD
	gitCommit     string // git rev-parse HEAD
)

func init() {
	flag.StringVar(&conf, "config", "", "Configuration file to use (default \"./clawiod.conf\")")
	flag.BoolVar(&version, "version", false, "Show version")
}

func main() {
	flag.Parse()

	if version {
		handleVersion()
	}

	cfg := config.New([]config.Source{defaul.New(), file.New(conf)})
	if err := cfg.LoadDirectives(); err != nil {
		log.Fatalf("cannot load configuration: %s", err)
	}

	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("cannot run daemon because: %s", err)
	}

	stopChan := d.TrapSignals()
	go d.Start()
	err = <-stopChan
	if err != nil {
		log.Fatalf("daemon finished execution with error: %s", err)
	} else {
		log.Info("daemon finished execution successfuly")
		os.Exit(0)
	}
}

func handleVersion() {
	// if gitTag is not empty we are on release build
	if gitTag != "" {
		fmt.Printf("%s %s commit:%s release-build\n", appName, gitNearestTag, gitCommit)
		os.Exit(0)
	}
	fmt.Printf("%s %s commit:%s dev-build\n", appName, gitNearestTag, gitCommit)
	os.Exit(0)
}
