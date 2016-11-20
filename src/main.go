package main

import (
	"fmt"
	"github.com/clawio/clawiod/src/grpc-servers/account"
	"github.com/clawio/clawiod/src/proto"
	"google.golang.org/grpc"
	"gopkg.in/ini.v1"
	"log"
	"net"
	"os"
)

var config = []byte(`
# Server Configuration
server.port = 1502
server.cpu = 100%
server.log.output = -
server.log.level =
server.log.maxsize =
server.log.maxage =
server.log.maxbackups =

# Session Configuration
sessionbackend.jwt.secret = jwt

# Account Service Configuration
userbackend = memory
userbackend.memory.users = labrador:test:Hugo Gonzalez Labrador,natalia:test:Natalia Iglesias Roqueiro
account.groupbackend = memory
account.groupbackend.memory.groups = admins:labrador,devs:natalia
account.ratelimit.authenticate = 100
account.ratelimit.validate = 100
account.ratelimit.whoami = 100

# MetaData Service Configuration
metadata.backend = local
metadata.ratelimit.stat = 100
metadata.ratelimit.mkdir = 100
metadata.ratelimit.rm = 100
metadata.ratelimit.mkhome = 100
metadata.ratelimit.mv = 100
metadata.local.datadir = /tmp
metadata.local.tempdir= /tmp


# Data Service Configuration
data.backend = local
data.local.datadir = /tmp
data.local.tempdir= /tmp
data.ratelimit.put = 100
data.ratelimit.get = 100
data.maxfilesize = 4096000
data.xs = md5
data.xsverify = false
data.eos.mgm = root://eosbackup.cern.ch
data.eos.userprefix = /eos/scratch/user

# WebDAV Service Configuration
webdav.datatarget = unix:/tmp/data.socket
webdav.metadatatarget = unix:/tmp/metadata.socket
webdav.ratelimit.propfind = 100
webdav.ratelimit.put = 100
webdav.ratelimit.get = 100
webdav.ratelimit.options = 100
webdav.ratelimit.head = 100
webdav.ratelimit.move = 100
webdav.ratelimit.copy = 100
webdav.ratelimit.lock = 100
webdav.ratelimit.unlock = 100

# OwnCloud Service Configuration
ocwebdav.datatarget = unix:/tmp/data.socket
ocwebdav.metadata.target = unix:/tmp/metadata.socket
ocwebdav.chunksdir = /tmp/chunks/
ocwebdav.chunkstemp = /tmp/chunks.temp/
`)

func main() {
	cfg, err := ini.Load(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	keys := cfg.Section("").Keys()
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key.Name(), key.Value())
	}

	accountServer, err := account.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Section("").Key("server.port").MustInt64(1502)))
	grpcServer := grpc.NewServer()
	proto.RegisterAccountServer(grpcServer, accountServer)
	grpcServer.Serve(lis)
}
