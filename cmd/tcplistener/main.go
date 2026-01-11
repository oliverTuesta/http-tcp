package main

import (
	"fmt"
	"log"
	"net"

	"github.com/oliverTuesta/http-tcp/internal/request"
)



const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatal("error", err)
	}	

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", err)
		}	

		fmt.Println("connection accepted")

		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", err)
		}

		fmt.Println("Request line:")	
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)	
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)	
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)	
		fmt.Println("Headers:")	
		for k, v := range request.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
	}

}

