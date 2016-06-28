package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	"github.com/clawio/clawiod/config/file"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/daemon"
)

const appName = "clawiod"

var log = logrus.WithField("module", "main")

// Flags that control program flow or startup
var (
	conf       string
	showconfig bool
	version    bool
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
	flag.BoolVar(&showconfig, "showconfig", false, "Show loaded configuration")
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

	if showconfig {
		handleShowConfig(cfg)
	}

	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("cannot run clawid daemon because: %s", err)
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

func handleShowConfig(cfg *config.Config) {
	dirs := cfg.GetDirectives()
	data, err := json.MarshalIndent(dirs, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
	os.Exit(0)
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

func handleCPU(cfg *config.Config) {
	cpu := cfg.GetDirectives().Server.CPU
	// Set CPU capacity
	err := setCPU(cpu)
	if err != nil {
		log.Fatal("Cannot tweak CPU: ", err)
	}
}

// setCPU parses string cpu and sets GOMAXPROCS
// according to its value. It accepts either
// a number (e.g. 3) or a percent (e.g. 50%).
func setCPU(cpu string) error {
	var numCPU int

	availCPU := runtime.NumCPU()

	if strings.HasSuffix(cpu, "%") {
		// Percent
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
	} else {
		// Number
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	runtime.GOMAXPROCS(numCPU)
	return nil
}
