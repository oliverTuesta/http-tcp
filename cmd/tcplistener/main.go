package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <- chan string {
	line := ""

	out := make(chan string)

	go func() {

		defer close(out)
		defer f.Close()

		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err != nil {
				break	
			}

			data = data[:n]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				line += string(data[:i])
				out <- line
				data = data[i+1:]
				line = ""
			}

			line += string(data)
		}

		if len(line) > 0 {
			out <- line
		}
	}()

	return out
}

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

		for line := range getLinesChannel(conn) {
			fmt.Printf("%s\n", line)
		}
	}

}

