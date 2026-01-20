package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/oliverTuesta/http-tcp/internal/headers"
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

func respond400() *server.HandlerError {
	return &server.HandlerError{
		StatusCode: response.StatusBadRequest,
		Message:    createBody(response.StatusBadRequest, "Bad Request", "Your request honestly kinda sucked"),
	}
}

func respond500() *server.HandlerError {
	return &server.HandlerError{
		StatusCode: response.StatusInternalServerError,
		Message:    createBody(response.StatusInternalServerError, "Internal Server Error", "Okay, you know what? This one is on me."),
	}
}

var VIDEO_PATH string = "assets/vim.mp4"

func main() {
	handler := server.Handler(func(w *response.Writer, req *request.Request) *server.HandlerError {
		target := req.RequestLine.RequestTarget

		if target == "/yourproblem" {
			return respond400()
		} else if target == "/myproblem" {
			return respond500()
		} else if strings.HasPrefix(target, "/httpbin/stream") {
			url := "https://httpbin.org/stream/" + target[len("/httpbin/stream/"):]
			res, err := http.Get(url)
			if err != nil {
				return respond500()	
			}

			hs := response.GetDefaultHeaders(0)
			hs.Delete("content-length")
			hs.Set("content-type", "text/plain")
			hs.Set("Transfer-Encoding", "chunked")
			hs.Delete("connection")
			w.WriteStatusLine(response.StatusOk)
			err = w.WriteHeaders(hs)
			if err != nil {
				log.Println("write error:", err)
			}

			hasher := sha256.New()
			totalLen := 0

			for {
				buffer := make([]byte, 1024)
				n, err := res.Body.Read(buffer)
				if err != nil {
					break
				}
				chunk := buffer[:n]

				hasher.Write(chunk)
				totalLen += n
				w.WriteChunkedBody(chunk)
			}
			w.WriteChunkedBodyDone()

			trailers := headers.NewHeaders()
			trailers.Set(
				"X-Content-SHA256",
				fmt.Sprintf("%x", hasher.Sum(nil)),
			)
			trailers.Set(
				"X-Content-Length",
				strconv.Itoa(totalLen),
			)
			if err := w.WriteTrailers(trailers); err != nil {
				log.Println("trailer write error:", err)
			}
		} else if target == "/video" {
			file, err := os.ReadFile(VIDEO_PATH)
			if err != nil {
				return respond500()	
			}

			hs := response.GetDefaultHeaders(len(file))
			hs.Set("content-type", "video/mp4")
			w.WriteStatusLine(response.StatusOk)
			err = w.WriteHeaders(hs)
			if err != nil {
				log.Println("write error:", err)
			}
			w.WriteBody(file)

		} else {
			body := createBody(response.StatusOk, "Success!", "Your request was an absolute banger.")
			hs := response.GetDefaultHeaders(len(body))
			w.WriteStatusLine(response.StatusOk)
			err := w.WriteHeaders(hs)
			if err != nil {
				log.Println("write error:", err)
			}
			w.WriteBody(body)
		}

		return nil
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
