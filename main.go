// ClawIO - Scalable Sync and Share
//
// Copyright (C) 2015  Hugo González Labrador <clawio@hugo.labkode.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. See file COPYNG.

package main

import (
	"flag"
	"fmt"
	"os"

	apidisp "github.com/clawio/clawiod/pkg/api/dispatcher"
	apiauth "github.com/clawio/clawiod/pkg/api/providers/auth"
	apifile "github.com/clawio/clawiod/pkg/api/providers/file"
	apistatic "github.com/clawio/clawiod/pkg/api/providers/static"
	apiwebdav "github.com/clawio/clawiod/pkg/api/providers/webdav"

	"github.com/clawio/clawiod/pkg/apiserver"

	authdisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	authfile "github.com/clawio/clawiod/pkg/auth/providers/file"

	storagedisp "github.com/clawio/clawiod/pkg/storage/dispatcher"
	storagelocal "github.com/clawio/clawiod/pkg/storage/providers/local"

	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	"github.com/clawio/clawiod/pkg/pidfile"
	"github.com/clawio/clawiod/pkg/signaler"
)

func main() {

	/*********************************************
	 *** 1. Parse CLI flags   ********************
	 *********************************************/
	flags := struct {
		pidFile string // the pidfile that will be used by the daemon
		cfg     string // the config that will be used by the daemon
		pc      bool   // if true prints the default config file
	}{}
	flag.StringVar(&flags.pidFile, "p", "", "PID file")
	flag.StringVar(&flags.cfg, "c", "", "Configuration file")
	flag.BoolVar(&flags.pc, "pc", false, "Prints the default configuration file")
	flag.Parse()

	if flags.pc == true {
		cfg, err := config.Default()
		if err != nil {
			fmt.Println("Cannot print default configuration: ", err)
			os.Exit(1)
		}
		fmt.Println(cfg)
		os.Exit(0)
	}
	/*********************************************
	 *** 2. Create PID file   ********************
	 *********************************************/
	if flags.pidFile == "" {
		fmt.Println("Set pidfile with -p flag")
		os.Exit(1)
	}
	_, err := pidfile.New(flags.pidFile)
	if err != nil {
		fmt.Println("Cannot create PID file: ", err)
		os.Exit(1)
	}

	/************************************************
	 *** 3. Load configuration   ********************
	 ************************************************/
	if flags.cfg == "" {
		fmt.Println("Set configuration with -c flag")
		os.Exit(1)
	}
	cfg, err := config.New(flags.cfg)
	if err != nil {
		fmt.Println("Cannot load configuration: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 4. Connect to the syslog daemon *******
	 ******************************************/
	syslogWriter, err := logger.NewSyslogWriter("", "", cfg.GetDirectives().LogLevel)
	if err != nil {
		fmt.Println("Cannot connect to syslog: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 5. Create auth dispatcher       *******
	 ******************************************/
	fileAuthLog := logger.New(syslogWriter, cfg.GetDirectives().LogLevel, "FILEAUTH")
	fauth, err := authfile.New("fileauth", cfg, fileAuthLog)
	if err != nil {
		fmt.Println("Cannot create file auth provider: ", err)
		os.Exit(1)
	}
	adispLog := logger.New(syslogWriter, cfg.GetDirectives().LogLevel, "AUTHDISP")
	adisp := authdisp.New(cfg, adispLog)
	err = adisp.AddAuth(fauth) // add file auth strategy
	if err != nil {
		fmt.Println("Cannot add file auth provider to auth dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 6. Create storage dispatcher      *****
	 ******************************************/
	localStorageLog := logger.New(syslogWriter, cfg.GetDirectives().LogLevel, "LOCALSTORAGE")
	localStorage := storagelocal.New("local", cfg, localStorageLog)

	sdispLog := logger.New(syslogWriter, cfg.GetDirectives().LogLevel, "STORAGEDISP")
	sdisp := storagedisp.New(cfg, sdispLog)
	err = sdisp.AddStorage(localStorage)
	if err != nil {
		fmt.Println("Cannot add local storage to storage dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 7. Create API dispatcher aka router  **
	 ******************************************/
	apdisp := apidisp.New(cfg)

	if cfg.GetDirectives().AuthAPIEnabled == true {
		authAPI := apiauth.New(cfg.GetDirectives().AuthAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(authAPI)
		if err != nil {
			fmt.Println("Cannot add auth API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().WebDAVAPIEnabled {
		webdavAPI := apiwebdav.New(cfg.GetDirectives().WebDAVAPIID, cfg, adisp, sdisp)

		err = apdisp.AddAPI(webdavAPI)
		if err != nil {
			fmt.Println("Cannot add WebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().FileAPIEnabled == true {
		fileAPI := apifile.New(cfg.GetDirectives().FileAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(fileAPI)
		if err != nil {
			fmt.Println("Cannot add File API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().StaticAPIEnabled == true {
		staticAPI := apistatic.New(cfg.GetDirectives().StaticAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(staticAPI)
		if err != nil {
			fmt.Println("Cannot add Static API to API dispatcher: ", err)
			os.Exit(1)
		}
	}
	/***************************************************
	 *** 8. Start HTTP/HTTPS Server ********************
	 ***************************************************/
	srv := apiserver.New(cfg, syslogWriter, apdisp, adisp, sdisp)
	go func() {
		err = srv.Start()
		if err != nil {
			fmt.Println("Cannot start HTTP/HTTPS API server: ", err)
			os.Exit(1)
		}
	}()

	/***************************************************
	 *** 9. Listen to OS signals to control the daemon *
	 ***************************************************/
	sig := signaler.New(cfg, srv)
	endc := sig.Start()
	<-endc
	os.Exit(0)
}
