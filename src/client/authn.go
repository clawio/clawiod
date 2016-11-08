package main

import (
	"log"
	"os"
	"google.golang.org/grpc"
	"github.com/clawio/clawiod/src/proto"
	"context"
)

func tokenRequest(con *grpc.ClientConn) {
	client := proto.NewAuthNClient(con)
	tokenRequest := &proto.TokenRequest{}
	tokenRequest.Username = tokenrequsername
	tokenRequest.Password = tokenrequestpassword
	tokenRequest.Opaque = tokenrequestopaque
	tokenResponse, err := client.Token(context.Background(), tokenRequest)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	debugCall(tokenRequest, tokenResponse)
}

func pingRequest(con *grpc.ClientConn) {
	client := proto.NewAuthNClient(con)
	pingRequest := &proto.PingRequest{}
	pingRequest.Token = pingrequesttoken
	pingResponse, err := client.Ping(context.Background(),pingRequest)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	debugCall(pingRequest, pingResponse)
}
