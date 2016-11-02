package main

import (
	"google.golang.org/grpc"
	"github.com/clawio/clawiod/src/proto"
	"os"
	"log"
	"context"
	"fmt"
	"flag"
)

var port int
var method string
var debug bool

var tokenrequsername string
var tokenrequestpassword string
var tokenrequestopaque string

func init() {
	flag.IntVar(&port, "port", 1502, "port")
	flag.StringVar(&method, "method", "", "method to be call")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.StringVar(&tokenrequsername, "tokenrequestusername", "", "token request username")
	flag.StringVar(&tokenrequestpassword, "tokenrequestpassword", "", "token request password")
	flag.StringVar(&tokenrequestopaque, "tokenrequestopaque", "", "token request opaque")
	flag.Parse()
}


func main() {
	con, err := grpc.Dial("localhost:1502", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	switch method {
	case "token":
		tokenRequest(con)
	case "ping":
		pingRequest(con)
	}
}

func debugCall(req interface{}, res interface{}) {
	if debug {
		fmt.Println(">>>>>>>")
		fmt.Println(req)
		fmt.Println("=======")
		fmt.Println(res)
		fmt.Println("<<<<<<<")
	} else {
		fmt.Println(res)
	}
}

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
	pingRequest.Token = "am83089qn jqegj77748nuuq98rnnny756"
	pingResponse, err := client.Ping(context.Background(),pingRequest)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	debugCall(pingRequest, pingResponse)
}
