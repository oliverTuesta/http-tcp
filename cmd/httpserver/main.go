package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/oliverTuesta/http-tcp/internal/request"
	"github.com/oliverTuesta/http-tcp/internal/response"
	"github.com/oliverTuesta/http-tcp/internal/server"
)

const port = 42069

func main() {

	handler := server.Handler(func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message: []byte("Your problem is not my problem\n"),
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message: []byte("Woopsie, my bad\n"),
			}
		default:
			w.Write([]byte("All good, frfr\n"))
			return nil
		}
	})

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
