package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {

	_, err := net.ResolveUDPAddr("udp", port)

	if err != nil {
		log.Fatal("error", err)
	}

	reader := bufio.NewReader(os.Stdin)

	dial, err := net.ResolveUDPAddr("udp", ":42069")

	if err != nil {
		log.Fatal("error", err)
	}


	conn, err := net.DialUDP("udp", nil, dial)

	if err != nil {
		log.Fatal("error", err)
	}

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("error", err)
		}

		conn.Write([]byte(input))
	}
}


