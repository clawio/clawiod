package main

import (
	"fmt"
	"github.com/clawio/clawiod/src/grpc-servers/authn"
	"github.com/clawio/clawiod/src/proto"
	"google.golang.org/grpc"
	"gopkg.in/ini.v1"
	"log"
	"net"
	"os"
)

var config = []byte(`
[server]
port=1502

[authn]
secret=supersecret
`)

func main() {
	cfg, err := ini.Load(config)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	authnServer := authn.New(cfg)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Section("server").Key("port").MustInt64(1502)))
	grpcServer := grpc.NewServer()
	proto.RegisterAuthNServer(grpcServer, authnServer)
	grpcServer.Serve(lis)
}
