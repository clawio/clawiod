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
	apiauth "github.com/clawio/clawiod/pkg/api/auth"
	apiocwebdav "github.com/clawio/clawiod/pkg/api/ocwebdav"
	apidisp "github.com/clawio/clawiod/pkg/api/pat"
	apistatic "github.com/clawio/clawiod/pkg/api/static"
	apistorage "github.com/clawio/clawiod/pkg/api/storage"
	apiwebdav "github.com/clawio/clawiod/pkg/api/webdav"
	authfile "github.com/clawio/clawiod/pkg/auth/file"
	authdisp "github.com/clawio/clawiod/pkg/auth/pat"
	config "github.com/clawio/clawiod/pkg/config/file"
	apiserver "github.com/clawio/clawiod/pkg/httpserver/api"
	logger "github.com/clawio/clawiod/pkg/logger/logrus"
	"github.com/clawio/clawiod/pkg/storage"
	//"github.com/clawio/clawiod/pkg/pidfile"
	"github.com/clawio/clawiod/pkg/signaler"
	storagelocal "github.com/clawio/clawiod/pkg/storage/local"
	storagedisp "github.com/clawio/clawiod/pkg/storage/pat"
	storageroot "github.com/clawio/clawiod/pkg/storage/root"
	"os"
)

// The version of the daemon.
const VERSION = "0.0.7"

type flags struct {
	cfg     string // the config that will be used by the daemon
	version bool
}

func parseFlags() *flags {
	flgs := &flags{}
	flag.StringVar(&flgs.cfg, "config", "", "load configuration from `file`")
	flag.BoolVar(&flgs.version, "version", false, "print the version")
	flag.Parse()
}

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
	flags := parseFlags()
	if flags.version == true {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if flags.cfg == "" {
		fmt.Fprintln(os.Stderr, "Set configuration file with -config flag")
		fmt.Fprintln(os.Stderr, "Run clawiod --help to obtain more information")
		os.Exit(1)
	}

	/************************************************
	 *** 2. Load configuration   ********************
	 ************************************************/
	cfg, err := config.New(flags.cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot load configuration: ", err)
		os.Exit(1)
	}


	appLogFd, err := os.OpenFile(cfg.GetDirectives().LogAppFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open app log file: ", err)
		os.Exit(1)
	}

	reqLogFd, err := os.OpenFile(cfg.GetDirectives().LogReqFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open req log file: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 5. Create auth pat       *******
	 ******************************************/
	fauth, err := authfile.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create file auth provider: ", err)
		os.Exit(1)
	}

	adisp := authdisp.New(cfg)

	err = adisp.AddAuthType(fauth) // add file auth strategy
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add file auth provider to auth pat: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 6. Create storage pat      *****
	 ******************************************/
	localStorage, err := storagelocal.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create local storage: ", err.Error())
		os.Exit(1)
	}

	// The storage prefix for root storage must be ALWAYS the empty string. This is the only way to get
	// OC sync clients connect to ClawIO skipping folder configuration.
	sts := []storage.Storage{localStorage}
	rootStorage := storageroot.New(cfg, sts)

	sdisp := storagedisp.New(cfg)

	err = sdisp.AddStorage(localStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add local storage to storage pat: ", err)
		os.Exit(1)
	}
	err = sdisp.AddStorage(rootStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add root storage to storage pat: ", err)
		os.Exit(1)
	}

	/***************************************************
	 *** 8. Start HTTP/HTTPS Server ********************
	 ***************************************************/
	serverParams := &apiserver.NewParams{}
	
	srv, err := apiserver.New(appLogFd, reqLogFd, apdisp, adisp, sdisp, cfg)
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
	signalerLog, err := logger.New(appLogFd, "signaler", cfg)
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr		os.Exit(1)
		}
	}
	sig := signaler.New(srv, cfg, signalerLog)
	endc := sig.Start()
	<-endc
	os.Exit(0)
}
