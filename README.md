# HTTP from TCP in Go

This repository contains the project built while completing the **Learn HTTP Protocol in Go** course on boot.dev.  
The goal of the project is to understand how HTTP works under the hood by implementing it directly on top of TCP, without using Go’s `net/http` server abstractions.

The project covers request parsing, header handling, response writing, chunked transfer encoding, trailers, and basic TCP/UDP networking.

## Project Structure

The main entry points live under `cmd/`:

- **cmd/httpserver**  
  A full HTTP server implemented on top of raw TCP. It handles requests, writes responses manually, supports chunked responses, trailers, and serves a video file.

- **cmd/tcplistener**  
  A simple TCP server that listens for connections and prints the parsed HTTP request (request line, headers, and body).

- **cmd/updsender**  
  A small UDP client used to send messages to a UDP address for networking experimentation.

## Running the Projects

### HTTP Server

Start the custom HTTP server:
```bash
go run cmd/httpserver/main.go
```
Test basic responses:
```bash
curl http://localhost:42069
curl http://localhost:42069/yourproblem
curl http://localhost:42069/myproblem
```

Test chunked streaming with trailers (proxying httpbin):
```
curl -v http://localhost:42069/httpbin/stream/5
```

#### Video Endpoint
The /video endpoint serves an MP4 file. First, download the video asset:
```
mkdir assets
curl -o assets/vim.mp4 https://storage.googleapis.com/qvault-webapp-dynamic-assets/lesson_videos/vim-vs-neovim-prime.mp4
```

Then request it:
```bash
curl -v http://localhost:42069/video --output video.mp4
```

### TCP Listener

Run the TCP listener:
```bash
go run cmd/tcplistener/main.go
```

Send a raw HTTP request using netcat:
```
printf "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n" | nc localhost 42069
```

The server will print the parsed request to stdout.

### UDP Sender

Run the UDP sender:
```bash
go run cmd/updsender/main.go
```
