package server

import (
	"bytes"
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
	Message    []byte
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("error:", err)
		return
	}

	var body bytes.Buffer

	handlerError := s.handler(&body, req)
	if handlerError != nil {
		WriteHandlerError(conn, handlerError)
	} else {
		headers := response.GetDefaultHeaders(body.Len())

		response.WriteStatusLine(conn, response.StatusOk)

		err := response.WriteHeaders(conn, headers)
		if err != nil {
			log.Println("write error:", err)
		}

		if _, err := io.Copy(conn, &body); err != nil {
			log.Println("write error:", err)
		}
	}

}

func WriteHandlerError(w io.Writer, handlerError *HandlerError) {
	response.WriteStatusLine(w, handlerError.StatusCode)
	headers := response.GetDefaultHeaders(len(handlerError.Message))
	err := response.WriteHeaders(w, headers)
	if err != nil {
		log.Println("write error:", err)
	}
	err = response.WriteBody(w, handlerError.Message)
	if err != nil {
		log.Println("write error:", err)
	}

}
