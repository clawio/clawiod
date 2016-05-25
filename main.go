package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/clawio/clawiod/config"
	"github.com/clawio/clawiod/config/default"
	"github.com/clawio/clawiod/config/file"
	"github.com/clawio/clawiod/daemon"
	"gopkg.in/natefinch/lumberjack.v2"
)

const appName = "ClawIO"

var log = logrus.WithField("module", "main")

// Flags that control program flow or startup
var (
	conf        string
	cpu         string
	port        int
	applogfile  string
	httplogfile string
	version     bool
)

// Build information obtained with the help of -ldflags
var (
	appVersion = "(untracked dev build)" // inferred at startup
	devBuild   = true                    // inferred at startup

	buildDate        string // date -u
	gitTag           string // git describe --exact-match HEAD 2> /dev/null
	gitNearestTag    string // git describe --abbrev=0 --tags HEAD
	gitCommit        string // git rev-parse HEAD
	gitShortStat     string // git diff-index --shortstat
	gitFilesModified string // git diff-index --name-only HEAD
)

func init() {
	flag.StringVar(&conf, "conf", "", "Configuration file to use (default \"./clawio.conf\")")
	flag.StringVar(&cpu, "cpu", "100%", "CPU capacity")
	flag.StringVar(&applogfile, "applogfile", "stdout", "File to log application data")
	flag.StringVar(&httplogfile, "httplogfile", "stdout", "File to log HTTP requests")
	flag.BoolVar(&version, "version", false, "Show version")
	flag.IntVar(&port, "port", 1502, "Port to listen for requests")
}

func main() {
	flag.Parse()
	configureLogger(applogfile)

	if version {
		handleVersion()
	}

	handleCPU()

	log.Info("cli flags parsed")
	printFlags()

	log.Info("will load configuration")
	cfg := config.New([]config.ConfigSource{defaul.New(), file.New(conf)})
	if err := cfg.LoadDirectives(); err != nil {
		log.Fatalf("cannot load configuration: %s", err)
	}
	log.Info("configuration loaded")
	directives := cfg.GetDirectives()
	configureLogger(directives.Server.AppLog)

	d, err := daemon.New(cfg)
	if err != nil {
		log.Fatalf("cannot instantiate the daemon: %s", err)
	}
	stopChan := d.TrapSignals()
	go d.Start()
	err = <-stopChan
	if err != nil {
		log.Fatal("daemon finished execution with error: %s", err)
	} else {
		log.Info("daemon finished execution successfuly")
		os.Exit(0)
	}
}

func printFlags() {
	log.WithField("flagkey", "conf").WithField("flagval", conf).Info("flag detail")
	log.WithField("flagkey", "cpu").WithField("flagval", cpu).Info("flag detail")
	log.WithField("flagkey", "applogfile").WithField("flagval", applogfile).Info("flag detail")
	log.WithField("flagkey", "httplogfile").WithField("flagval", httplogfile).Info("flag detail")
	log.WithField("flagkey", "port").WithField("flagval", port).Info("flag detail")
}

func configureLogger(applogfile string) {

	switch applogfile {
	case "stdout":
		log.Logger.Out = os.Stdout
	case "stderr":
		log.Logger.Out = os.Stderr
	case "":
		log.Logger.Out = ioutil.Discard
	default:
		log.Logger.Out = &lumberjack.Logger{
			Filename:   applogfile,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		}
	}
}

func handleVersion() {
	fmt.Printf("%s %s\n", appName, appVersion)
	if devBuild && gitShortStat != "" {
		fmt.Printf("%s\n%s\n", gitShortStat, gitFilesModified)
	}
	os.Exit(0)
}

func handleCPU() {
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
func redacted(v string) string {
	length := len(v)
	if length == 0 {
		return ""
	}
	if length == 1 {
		return "X"
	}
	half := length / 2
	right := v[half:]
	hidden := strings.Repeat("X", 10)
	return strings.Join([]string{hidden, right}, "")
}
