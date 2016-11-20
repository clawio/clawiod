package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/clawio/clawiod/src/proto"
	"google.golang.org/grpc"
	"log"
	"os"
)

var address string
var method string
var debug bool
var tojson bool
var getticketrequestcredentialsprotocol int
var getticketrequestcredentialscredentials string
var whoamiticket string
var listmethods bool

func init() {
	flag.StringVar(&address, "address", "localhost:1502", "address of remote server")
	flag.StringVar(&method, "method", "", "method to be call")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&tojson, "tojson", false, "encode req and response as JSON")
	flag.BoolVar(&listmethods, "listmethods", false, "list available methods")
	flag.IntVar(&getticketrequestcredentialsprotocol, "getticketrequestcredentialsprotocol", 0, "protocol (0 => Basic, 1 => KRB")
	flag.StringVar(&getticketrequestcredentialscredentials, "getticketrequestcredentialscredentials", "", "credentials")
	flag.StringVar(&whoamiticket, "whoamiticket", "", "The session ticket")
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if listmethods {
		fmt.Println("GetTicket")
		os.Exit(0)
	}

	switch method {
	case "Authenticate":
		getTicketRequest(con)
	case "Whoami":
		whoami(con)
	default:
		log.Fatal("method doesn't exist")
		os.Exit(1)
	}
}

func debugCall(req interface{}, res interface{}) {
	if tojson {
		req, _ = json.Marshal(req)
		res, _ = json.Marshal(res)
		req = string(req.([]byte))
		res = string(res.([]byte))
	}
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

func getTicketRequest(con *grpc.ClientConn) {
	secCredentials := &proto.SecCredentials{}
	secCredentials.Protocol = proto.ProtocolType(getticketrequestcredentialsprotocol)
	secCredentials.Credentials = getticketrequestcredentialscredentials

	ticketRequest := &proto.AuthenticateRequest{}
	ticketRequest.SecCredentials = secCredentials

	client := proto.NewAccountClient(con)
	ticketResponse, err := client.Authenticate(context.Background(), ticketRequest)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	debugCall(ticketRequest, ticketResponse)
}

func whoami(con *grpc.ClientConn) {
	req := &proto.WhoamiRequest{Ticket: whoamiticket}
	client := proto.NewAccountClient(con)
	res, err := client.Whoami(context.Background(), req)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	debugCall(req, res)
}
