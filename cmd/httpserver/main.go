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

func createBody(status response.StatusCode, title string, description string) []byte {
	return []byte(
		`<html>
		<head>
		<title>` + string(status) + `</title>
		</head>
		<body>
		<h1>` + title + `</h1>
		<p>` + description + `</p>
		</body>
		</html>`,
	)
}

func main() {
	handler := server.Handler(func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			return &server.HandlerError{
				StatusCode: response.StatusBadRequest,
				Message:    createBody(response.StatusBadRequest, "Bad Request", "Your request honestly kinda sucked"),
			}
		case "/myproblem":
			return &server.HandlerError{
				StatusCode: response.StatusInternalServerError,
				Message:    createBody(response.StatusInternalServerError, "Internal Server Error", "Okay, you know what? This one is on me."),
			}
		default:
			w.Write(createBody(response.StatusOk, "Success!", "Your request was an absolute banger."))
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
