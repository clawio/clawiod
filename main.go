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
	apidisp "github.com/clawio/clawiod/pkg/api/dispatcher"
	apiauth "github.com/clawio/clawiod/pkg/api/providers/auth"
	apistatic "github.com/clawio/clawiod/pkg/api/providers/static"
	apistorage "github.com/clawio/clawiod/pkg/api/providers/storage"
	apiwebdav "github.com/clawio/clawiod/pkg/api/providers/webdav"
	"github.com/clawio/clawiod/pkg/apiserver"
	authdisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	authfile "github.com/clawio/clawiod/pkg/auth/providers/file"
	"github.com/clawio/clawiod/pkg/config"
	"github.com/clawio/clawiod/pkg/logger"
	//"github.com/clawio/clawiod/pkg/pidfile"
	"github.com/clawio/clawiod/pkg/signaler"
	storagedisp "github.com/clawio/clawiod/pkg/storage/dispatcher"
	storagelocal "github.com/clawio/clawiod/pkg/storage/providers/local"
	"os"
)

func main() {

	// The daemon MUST run as non-root user to avoid security holes.
	// Linux threads are not POSIX compliant so the setuid sycall just apply to the actual thread. This
	// makes setuid not safe. See https://github.com/golang/go/issues/1435
	// There are two options to listen in a port < 1024 (privileged ports)
	// I) Use Linux capabilities: sudo setcap cap_net_bind_service=+ep clawiod
	// II) Use a reverse proxy like NGINX or lighthttpd that listen on 80 and forwards to daemon on port > 1024

	/*********************************************
	 *** 1. Parse CLI flags   ********************
	 *********************************************/
	flags := struct {
		//pidFile string // the pidfile that will be used by the daemon
		cfg string // the config that will be used by the daemon
	}{}
	//flag.StringVar(&flags.pidFile, "pid", "", "The pid file")
	flag.StringVar(&flags.cfg, "config", "", "the configuration file")
	flag.Parse()
	if flags.cfg == "" /*&& flags.pidFile == ""*/ {
		flag.PrintDefaults()
		os.Exit(1)
	}

	/*********************************************
	 *** 2. Create PID file   ********************
	 *********************************************/
	/*if flags.pidFile == "" {
		fmt.Fprintln(os.Stderr,"Set pidfile with -pid flag")
		os.Exit(1)
	}
	_, err := pidfile.New(flags.pidFile)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Cannot create PID file: ", err)
		os.Exit(1)
	}*/

	/************************************************
	 *** 3. Load configuration   ********************
	 ************************************************/
	if flags.cfg == "" {
		fmt.Fprintln(os.Stderr, "Set configuration file with -config flag")
		os.Exit(1)
	}
	cfg, err := config.New(flags.cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load configuration: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 4. Connect to the syslog daemon *******
	 ******************************************/
	/*syslogWriter, err := logger.NewSyslogWriter("", "", cfg.GetDirectives().LogLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Cannot connect to syslog: ", err)
		os.Exit(1)
	}*/

	appLogWriter, err := os.OpenFile(cfg.GetDirectives().LogAppFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open app log file: ", err)
		os.Exit(1)
	}

	reqLogWriter, err := os.OpenFile(cfg.GetDirectives().LogReqFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open req log file: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 5. Create auth dispatcher       *******
	 ******************************************/
	fileAuthLog := logger.New(appLogWriter, "FILEAUTH")
	fauth, err := authfile.New("fileauth", cfg, fileAuthLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create file auth provider: ", err)
		os.Exit(1)
	}
	adispLog := logger.New(appLogWriter, "AUTHDISP")
	adisp := authdisp.New(cfg, adispLog)
	err = adisp.AddAuthenticationstrategy(fauth) // add file auth strategy
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add file auth provider to auth dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 6. Create storage dispatcher      *****
	 ******************************************/
	localStorageLog := logger.New(appLogWriter, "LOCALSTORAGE")
	localStorage := storagelocal.New("local", cfg, localStorageLog)

	sdispLog := logger.New(appLogWriter, "STORAGEDISP")
	sdisp := storagedisp.New(cfg, sdispLog)
	err = sdisp.AddStorage(localStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add local storage to storage dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 7. Create API dispatcher             **
	 ******************************************/
	apdisp := apidisp.New(cfg)

	if cfg.GetDirectives().AuthAPIEnabled == true {
		authAPI := apiauth.New(cfg.GetDirectives().AuthAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(authAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add auth API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().WebDAVAPIEnabled {
		webdavAPI := apiwebdav.New(cfg.GetDirectives().WebDAVAPIID, cfg, adisp, sdisp)

		err = apdisp.AddAPI(webdavAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add WebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().StorageAPIEnabled == true {
		storageAPI := apistorage.New(cfg.GetDirectives().StorageAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(storageAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Storage API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().StaticAPIEnabled == true {
		staticAPI := apistatic.New(cfg.GetDirectives().StaticAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(staticAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Static API to API dispatcher: ", err)
			os.Exit(1)
		}
	}
	/***************************************************
	 *** 8. Start HTTP/HTTPS Server ********************
	 ***************************************************/
	srv := apiserver.New(cfg, appLogWriter, reqLogWriter, apdisp, adisp, sdisp)
	go func() {
		err = srv.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot start HTTP/HTTPS API server: ", err)
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
