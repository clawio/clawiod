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

	idmjson "github.com/clawio/clawiod/pkg/auth/file"
	idmpat "github.com/clawio/clawiod/pkg/auth/pat"
	config "github.com/clawio/clawiod/pkg/config/file"
	server "github.com/clawio/clawiod/pkg/httpserver"
	"github.com/clawio/clawiod/pkg/signaler"
	"github.com/clawio/clawiod/pkg/storage"
	strglocal "github.com/clawio/clawiod/pkg/storage/local"
	strgpat "github.com/clawio/clawiod/pkg/storage/pat"
	storageroot "github.com/clawio/clawiod/pkg/storage/root"

	apiauth "github.com/clawio/clawiod/pkg/api/auth"
	apiocwebdav "github.com/clawio/clawiod/pkg/api/ocwebdav"
	apipat "github.com/clawio/clawiod/pkg/api/pat"
	apistatic "github.com/clawio/clawiod/pkg/api/static"
	apistorage "github.com/clawio/clawiod/pkg/api/storage"
	apiwebdav "github.com/clawio/clawiod/pkg/api/webdav"
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
	return flgs
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

	/************************************************
	 *** 3. Create log writers   ********************
	 ************************************************/
	appWriter, err := os.OpenFile(cfg.GetDirectives().LogAppFile,
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open app log file: ", err)
		os.Exit(1)
	}

	reqWriter, err := os.OpenFile(cfg.GetDirectives().LogReqFile,
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot open req log file: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 4. Create identity manager dispatcher *
	 ******************************************/
	idmPat := idmpat.New(cfg)

	idmJSONParams := &idmjson.NewParams{}
	idmJSONParams.Config = cfg
	idmJSONParams.ID = "fileauth"

	idmJSON, err := idmjson.New(idmJSONParams)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create JSON idm: ", err)
		os.Exit(1)
	}

	err = idmPat.AddIDM(idmJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add JSON idm to idm pat: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 5. Create storage dispatcher      *****
	 ******************************************/
	strgPatParams := &strgpat.NewParams{}
	strgPatParams.Config = cfg

	strgPat := strgpat.New(strgPatParams)

	strgLocalParams := &strglocal.NewParams{}
	strgLocalParams.Config = cfg
	strgLocalParams.Prefix = "local"

	strgLocal, err := strglocal.New(strgLocalParams)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create local storage: ", err.Error())
		os.Exit(1)
	}

	// The storage prefix for root storage must be ALWAYS the empty string. This is the only way to get
	// OC sync clients connect to ClawIO skipping folder configuration.
	strgRootParams := &storageroot.NewParams{}
	strgRootParams.Config = cfg
	strgRootParams.Prefix = ""
	strgRootParams.Storages = []storage.Storage{strgLocal}

	strgRoot := storageroot.New(strgRootParams)

	addStorageParams := &strgpat.AddStorageParams{}
	addStorageParams.Storage = strgLocal
	err = strgPat.AddStorage(addStorageParams)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add local storage to storage pat: ", err)
		os.Exit(1)
	}
	addStorageParams.Storage = strgRoot
	err = strgPat.AddStorage(addStorageParams)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot add root storage to storage pat: ", err)
		os.Exit(1)
	}

	/******************************************
	 ** 6. Create API dispatcher          *****
	 ******************************************/
	apiPatParams := &apipat.NewParams{}
	apiPatParams.Config = cfg
	apiPat := apipat.New(apiPatParams)

	if cfg.GetDirectives().AuthAPIEnabled == true {
		apiAuthParams := &apiauth.NewParams{}
		apiAuthParams.Config = cfg

		apiAuth := apiauth.New(apiAuthParams)
		err = apiPat.AddAPI(apiAuth)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Auth API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().WebDAVAPIEnabled {
		apiWebDAVParams := &apiwebdav.NewParams{}
		apiWebDAVParams.Config = cfg

		webdavAPI := apiwebdav.New(apiWebDAVParams)
		err = apiPat.AddAPI(webdavAPI)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add WebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().OCWebDAVAPIEnabled {
		apiOCWebDAVParams := &apiocwebdav.NewParams{}
		apiOCWebDAVParams.Config = cfg

		apiOCWebDAV := apiocwebdav.New(apiOCWebDAVParams)
		err = apiPat.AddAPI(apiOCWebDAV)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add OCWebDAV API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().StorageAPIEnabled == true {
		apiStorageParams := &apistorage.NewParams{}
		apiStorageParams.Config = cfg

		apiStorage := apistorage.New(apiStorageParams)
		err = apiPat.AddAPI(apiStorage)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Storage API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	if cfg.GetDirectives().StaticAPIEnabled == true {
		apiStaticParams := &apistatic.NewParams{}
		apiStaticParams.Config = cfg

		apiStatic := apistatic.New(apiStaticParams)
		err = apiPat.AddAPI(apiStatic)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot add Static API to API dispatcher: ", err)
			os.Exit(1)
		}
	}

	/***************************************************
	 *** 8. Start HTTP/HTTPS Server ********************
	 ***************************************************/

	srvParams := &server.NewParams{}
	srvParams.AppLogWriter = appWriter
	srvParams.ReqLogWriter = reqWriter
	srvParams.Config = cfg
	srvParams.IDMPat = idmPat
	srvParams.StoragePat = strgPat

	srv, err := server.New(srvParams)
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
	sigParams := &signaler.NewParams{}
	sigParams.Config = cfg
	sigParams.Server = srv

	sig := signaler.New(sigParams)
	endc := sig.Start()

	host, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get hostname: ", err)
	}
	fmt.Fprintf(os.Stdout, "ClawIO Daemon listening on %s:%d", host, cfg.GetDirectives().Port)

	<-endc
	os.Exit(0)
}
