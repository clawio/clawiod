package main

import (
"google.golang.org/grpc"
"os"
"log"
"fmt"
"flag"
"encoding/json"
)

var address string
var method string
var debug bool
var jsonencoding bool

var tokenrequsername string
var tokenrequestpassword string
var tokenrequestopaque string

var pingrequesttoken string

func init() {
	flag.StringVar(&address, "address", "localhost:1502", "address of remote server")
	flag.StringVar(&method, "method", "", "method to be call")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&jsonencoding, "jsonencoding", false, "encode request and response as JSON")
	flag.StringVar(&tokenrequsername, "tokenrequestusername", "", "token request username")
	flag.StringVar(&tokenrequestpassword, "tokenrequestpassword", "", "token request password")
	flag.StringVar(&tokenrequestopaque, "tokenrequestopaque", "", "token request opaque")
	flag.StringVar(&pingrequesttoken, "pingrequesttoken", "", "ping request token")
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
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
	if jsonencoding {
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

