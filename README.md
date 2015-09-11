
# ClawIO Daemon
The official documentation of the project is hosted at  [clawio.github.io](http://clawio.github.io).

The ClawIO Daemon (clawiod) is the core component of the [ClawIO project](http://clawio.github.io).

The daemon runs a HTTP/HTTPS server that exposes a REST API to manage storages.

# Code documentation
| GoDoc
-------
|[![GoDoc](https://godoc.org/github.com/clawio/clawiod?status.svg)](https://godoc.org/github.com/clawio/clawiod/pkg)
# Build status
 Master        | Develop           
| ------------- |:-------------:
| [![Build Status](https://travis-ci.org/clawio/clawiod.svg?branch=master)](https://travis-ci.org/clawio/clawiod)|[![Build Status](https://travis-ci.org/clawio/clawiod.svg?branch=develop)](https://travis-ci.org/clawio/clawiod)

# Installation for developers
* [Install Go](https://golang.org/doc/install)
* [Configure your $GOPATH](https://golang.org/doc/code.html#Workspaces)
* `go get github.com/tools/godep`
* `mkdir -p $GOPATH/src/github.com/clawio`
* `cd $GOPATH/src/github.com/clawio`
* `git clone https://github.com/clawio/clawiod`
* `cd $GOPATH/src/github.com/clawio/clawiod`
* `godep restore`
* `go build`
* `./clawiod --help`
 

# Usage
```
clawiod --help
Usage of clawiod:
  -config configfilename
     use configfilename as the configuration file
  -version
    	print the version
```
To run the daemon you only need to specify the location of the configuration file and make sure you have permission to access it.

# How to control the daemon
The daemon is controlled by sending kernel signals to the daemon's process.

- Reload configuration.
```
kill -SIGHUP pid
```
- Hard shutdown of the server (**not recommended** because aborts ongoing requests)
```
kill -SIGTERM pid
kill -SIGINT  pid
```
- Graceful shutdown of the server (**recommended**). Stops accepting new requests and ongoing requests have the opportunity to finish within the timeout specified in the configuration
```
kill -SIGQUIT pid
```

# License
See COPYING file
