package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/clawio/lib/apiserver"
	"github.com/clawio/lib/config"
	"github.com/clawio/lib/pidfile"
	"github.com/clawio/lib/signaler"
)

const PRODUCTNAME = "clawio"

func main() {

	/*********************************************
	 *** 1. Parse CLI flags   ********************
	 *********************************************/
	flags := struct {
		pidFile string // the pidfile that will be used by the daemon
		cfg     string
	}{}
	flag.StringVar(&flags.pidFile, "pidfile", fmt.Sprintf("/var/run/%s.pid", PRODUCTNAME), "PID file location")
	flag.StringVar(&flags.cfg, "config", fmt.Sprintf("/etc/sysconfig/%s.conf", PRODUCTNAME), "Configuration file location")
	flag.Parse()

	/*********************************************
	 *** 2. Create PID file   ********************
	 *********************************************/
	_, err := pidfile.New(flags.pidFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/************************************************
	 *** 3. Load configuration   ********************
	 ************************************************/
	cfg, err := config.New(flags.cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	/***************************************************
	 *** 4. Start HTTP/HTTPS Server ********************
	 ***************************************************/
	srv := apiserver.New(cfg, nil)
	go srv.Start()

	/***************************************************
	 *** 5. Listen to OS signals to control the daemon *
	 ***************************************************/
	sig := signaler.New(cfg, srv)
	endc := sig.Start()
	<-endc
	fmt.Println("END")
	os.Exit(0)
}
