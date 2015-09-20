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
	apiocwebdav "github.com/clawio/clawiod/pkg/api/providers/ocwebdav"
	apistatic "github.com/clawio/clawiod/pkg/api/providers/static"
	apistorage "github.com/clawio/clawiod/pkg/api/providers/storage"
	apiwebdav "github.com/clawio/clawiod/pkg/api/providers/webdav"
	authdisp "github.com/clawio/clawiod/pkg/auth/dispatcher"
	authfile "github.com/clawio/clawiod/pkg/auth/providers/file"
	config "github.com/clawio/clawiod/pkg/config/file"
	apiserver "github.com/clawio/clawiod/pkg/httpserver/api"
	logger "github.com/clawio/clawiod/pkg/logger/logrus"
	"github.com/clawio/clawiod/pkg/storage"
	//"github.com/clawio/clawiod/pkg/pidfile"
	"github.com/clawio/clawiod/pkg/signaler/signalone"
	storagedisp "github.com/clawio/clawiod/pkg/storage/dispatcher"
	storagelocal "github.com/clawio/clawiod/pkg/storage/providers/local"
	storageroot "github.com/clawio/clawiod/pkg/storage/providers/root"
	"os"
)

// The version of the daemon.
const VERSION = "0.0.7"

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
		cfg     string // the config that will be used by the daemon
		version bool
	}{}
	//flag.StringVar(&flags.pidFile, "pid", "", "The pid file")
	flag.StringVar(&flags.cfg, "config", "", "use `configfilename` as the configuration file")
	flag.BoolVar(&flags.version, "version", false, "print the version")
	flag.Parse()
	if flags.version == true {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if flags.cfg == "" {
		fmt.Fprintln(os.Stderr, "Set configuration file with -config flag")
		fmt.Fprintln(os.Stderr, "Run clawiod --help to obtain more information")
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
	cfg, err := config.New(flags.cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load configuration: ", err)
		os.Exit(1)
	}
	directives, err := cfg.GetDirectives()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot read configuration directives: ", err)
		os.Exit(1)
	}
	/******************************************
	 ** 4. Connect to the syslog daemon *******
	 ******************************************/
	/*syslogWriter, err := logger.NewSyslogWriter("", "", directives.LogLevel)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Cannot connect to syslog: ", err)
		os.Exit(1)
	}*/

	appLogWriter, err := os.OpenFile(directives.LogAppFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open app log file: ", err)
		os.Exit(1)
	}

	reqLogWriter, err := os.OpenFile(directives.LogReqFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open req log file: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 5. Create auth dispatcher       *******
	 ******************************************/
	fileAuthLog, err := logger.New(appLogWriter, fmt.Sprintf("authid-%s", directives.FileAuthAuthID), cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create file auth logger: ", err.Error())
		os.Exit(1)
	}
	fauth, err := authfile.New(directives.FileAuthAuthID, cfg, fileAuthLog)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create file auth provider: ", err)
		os.Exit(1)
	}
	adispLog, err := logger.New(appLogWriter, "authdispatcher", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create auth dispatcher logger: ", err.Error())
		os.Exit(1)
	}
	adisp := authdisp.New(cfg, adispLog)
	err = adisp.AddAuthenticationstrategy(fauth) // add file auth strategy
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add file auth provider to auth dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 6. Create storage dispatcher      *****
	 ******************************************/
	localStorageLog, err := logger.New(appLogWriter, fmt.Sprintf("storage-%s", directives.LocalStoragePrefix), cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create local storage logger: ", err.Error())
		os.Exit(1)
	}
	localStorage := storagelocal.New(directives.LocalStoragePrefix, cfg, localStorageLog)

	// The storage prefix for root storage must be ALWAYS the empty string. This is the only way to get
	// OC sync clients connect to ClawIO skipping folder configuration.
	rootStorageLog, err := logger.New(appLogWriter, "storage-root", cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create root storage logger: ", err.Error())
		os.Exit(1)
	}
	sts := []storage.Storage{localStorage}
	rootStorage := storageroot.New("", sts, cfg, rootStorageLog)

	sdispLog, err := logger.New(appLogWriter, "storagedispatcher", cfg)
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot create storage dispatcher logger: ", err.Error())
			os.Exit(1)
		}
	}
	sdisp := storagedisp.New(cfg, sdispLog)
	err = sdisp.AddStorage(localStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add local storage to storage dispatcher: ", err)
		os.Exit(1)
	}
	err = sdisp.AddStorage(rootStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add root storage to storage dispatcher: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 7. Create API dispatcher             **
	 ******************************************/
	apiDispatcherLog, err := logger.New(appLogWriter, "apidispatcher", cfg)
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot create api dispatcher logger: ", err.Error())
			os.Exit(1)
		}
	}
	apdisp := apidisp.New(cfg, apiDispatcherLog)

	if directives.AuthAPIEnabled == true {
		apiAuthLog, err := logger.New(appLogWriter, "apiauth", cfg)
		if err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Cannot create api auth logger: ", err.Error())
				os.Exit(1)
			}
		}
		authAPI := apiauth.New(directives.AuthAPIID, adisp, sdisp, cfg, apiAuthLog)
		err = apdisp.AddAPI(authAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Auth API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if directives.WebDAVAPIEnabled {
		apiWebDAVLog, err := logger.New(appLogWriter, "apiwebdav", cfg)
		if err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Cannot create api webdav logger: ", err.Error())
				os.Exit(1)
			}
		}
		webdavAPI := apiwebdav.New(directives.WebDAVAPIID, adisp, sdisp, cfg, apiWebDAVLog)
		err = apdisp.AddAPI(webdavAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add WebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if directives.OCWebDAVAPIEnabled {
		apiOCWebDAVLog, err := logger.New(appLogWriter, "apiocwebdav", cfg)
		if err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Cannot create api ocwebdav logger: ", err.Error())
				os.Exit(1)
			}
		}
		ocwebdavAPI := apiocwebdav.New(directives.OCWebDAVAPIID, adisp, sdisp, cfg, apiOCWebDAVLog)
		err = apdisp.AddAPI(ocwebdavAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add OCWebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if directives.StorageAPIEnabled == true {
		apiStorageLog, err := logger.New(appLogWriter, "apistorage", cfg)
		if err != nil {
			if err != nil {
				fmt.Fprintln(os.Stderr, "Cannot create api storage logger: ", err.Error())
				os.Exit(1)
			}
		}
		storageAPI := apistorage.New(directives.StorageAPIID, adisp, sdisp, cfg, apiStorageLog)
		err = apdisp.AddAPI(storageAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Storage API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if directives.StaticAPIEnabled == true {
		staticAPI := apistatic.New(directives.StaticAPIID, cfg, adisp, sdisp)
		err = apdisp.AddAPI(staticAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Static API to API dispatcher: ", err)
			os.Exit(1)
		}
	}
	/***************************************************
	 *** 8. Start HTTP/HTTPS Server ********************
	 ***************************************************/
	srv, err := apiserver.New(appLogWriter, reqLogWriter, apdisp, adisp, sdisp, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create API server: ", err)
	}
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
	signalerLog, err := logger.New(appLogWriter, "signaler", cfg)
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot create signaler logger: ", err.Error())
			os.Exit(1)
		}
	}
	sig := signalone.New(srv, cfg, signalerLog)
	endc := sig.Start()
	<-endc
	os.Exit(0)
}
