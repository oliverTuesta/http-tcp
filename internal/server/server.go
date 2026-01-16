package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/oliverTuesta/http-tcp/internal/request"
	"github.com/oliverTuesta/http-tcp/internal/response"
)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, handler Handler) (*Server, error) {
	url := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", url)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println("Error accepting connection:", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {

	defer conn.Close()

	fmt.Println("Request line 0:")
	req, err := request.RequestFromReader(conn)
	fmt.Println("Request line 2:")
	if err != nil {
		log.Println("error:", err)
		return
	}

	// logs
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", req.RequestLine.Method)
	fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for k, v := range req.Headers {
		fmt.Printf("- %s: %s\n", k, v)
	}

	fmt.Println("Body:")
	fmt.Println(string(req.Body))
	//

	handlerError := s.handler(conn, req)
	if handlerError != nil {
		WriteHandlerError(conn, handlerError)
	} else {
		response.WriteStatusLine(conn, response.StatusOk)
		headers := response.GetDefaultHeaders(0)
		err := response.WriteHeaders(conn, headers)
		if err != nil {
			log.Println("write error:", err)
		}
	}

}

func WriteHandlerError(w io.Writer, handlerError *HandlerError) {
	response.WriteStatusLine(w, handlerError.StatusCode)
	headers := response.GetDefaultHeaders(0)
	err := response.WriteHeaders(w, headers)
	if err != nil {
		log.Println("write error:", err)
	}
}
