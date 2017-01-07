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
	"github.com/clawio/clawiod/config/etcd"
	"github.com/clawio/clawiod/daemon"
)

const appName = "clawiod"

var log = logrus.WithField("module", "main")

// Flags that control program flow or startup
var (
	conf             string
	showconfig       bool
	version          bool
	etcdconfurls     string
	etcdconfusername string
	etcdconfpassword string
	etcdconfkey      string
)

// Build information obtained with the help of -ldflags
var (
	buildDate     string // date -u
	gitTag        string // git describe --exact-match HEAD
	gitNearestTag string // git describe --abbrev=0 --tags HEAD
	gitCommit     string // git rev-parse HEAD
)

func init() {
	flag.StringVar(&conf, "config.file", "", "Configuration file to use (default \"./clawiod.conf\")")
	flag.StringVar(&etcdconfurls, "config.etcd.urls", "", "ETCD urls separated by comma")
	flag.StringVar(&etcdconfusername, "config.etcd.username", "", "ETCD username")
	flag.StringVar(&etcdconfpassword, "config.etcd.password", "", "ETCD password")
	flag.StringVar(&etcdconfkey, "config.etcd.key", "", "ETCD configuration key")
	flag.BoolVar(&showconfig, "showconfig", false, "Show loaded configuration")
	flag.BoolVar(&version, "version", false, "Show version")
}

func main() {
	flag.Parse()

	if version {
		handleVersion()
	}

	cfg := getConfig()
	if err := cfg.LoadDirectives(); err != nil {
		log.Fatalf("error loading configuration: %s", err)
	}

	if showconfig {
		handleShowConfig(cfg)
	}

	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("error running daemon: %s", err)
	}

	stopChan := d.TrapSignals()
	go d.Start()
	err = <-stopChan
	if err != nil {
		log.Fatalf("daemon stopped with error: %s", err)
	} else {
		log.Info("daemon stopped cleanly")
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
		fmt.Printf("%s %s commit:%s release-build %s\n", appName, gitNearestTag, gitCommit, buildDate)
		os.Exit(0)
	}
	fmt.Printf("%s %s commit:%s dev-build %s\n", appName, gitNearestTag, gitCommit, buildDate)
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

func getConfig() *config.Config {
	sources := []config.Source{defaul.New()}
	if etcdconfurls != "" {
		etcdSource, err := etcd.New(etcdconfurls, etcdconfkey, etcdconfusername, etcdconfpassword)
		if err != nil{
			log.Fatal("etcd configuration is bad: %s", err.Error())
		}
		sources = append(sources, etcdSource)
	} else {
		sources = append(sources, file.New(conf))
	}
	return config.New(sources)
}
