package server

import (
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

type Handler func(w *response.Writer, req *request.Request) *HandlerError

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

	writer := response.NewWriter(conn)

	handlerError := s.handler(writer, req)
	if handlerError != nil {
		WriteHandlerError(conn, handlerError)
	}
}

func WriteHandlerError(w io.Writer, handlerError *HandlerError) {
	writer := response.NewWriter(w)
	writer.WriteStatusLine(handlerError.StatusCode)
	headers := response.GetDefaultHeaders(len(handlerError.Message))
	err := writer.WriteHeaders(headers)
	if err != nil {
		log.Println("write error:", err)
	}
	_, err = writer.WriteBody(handlerError.Message)
	if err != nil {
		log.Println("write error:", err)
	}

}
