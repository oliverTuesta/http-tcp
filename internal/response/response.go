package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/oliverTuesta/http-tcp/internal/headers"
)

type StatusCode string

const (
	StatusOk                  StatusCode = "200 OK"
	StatusBadRequest          StatusCode = "400 Bad Request"
	StatusNotFound            StatusCode = "404 Not Found"
	StatusInternalServerError StatusCode = "500 Internal Server Error"
)

type writerState int

const (
	stateInit writerState = iota
	stateStatusWritten
	stateHeadersWritten
	stateBodyWritten
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w, state: stateInit}
}

var ERROR_INVALID_STATUS_CODE = fmt.Errorf("invalid status code")
var ERROR_INVALID_WRITE_ORDER = fmt.Errorf("invalid writer order")

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateInit {
		return ERROR_INVALID_WRITE_ORDER
	}

	switch statusCode {
	case StatusOk,
		StatusBadRequest,
		StatusNotFound,
		StatusInternalServerError:
	default:
		return ERROR_INVALID_STATUS_CODE
	}

	_, err := w.writer.Write([]byte("HTTP/1.1 " + string(statusCode) + "\r\n"))
	w.state = stateStatusWritten
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	var headers = headers.NewHeaders()
	headers.Set("content-length", strconv.Itoa(contentLen))
	headers.Set("connection", "close")
	headers.Set("content-type", "text/html")
	return headers
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != stateStatusWritten {
		return ERROR_INVALID_WRITE_ORDER
	}
	for key, value := range headers {
		_, err := w.writer.Write([]byte(key + ": " + value + "\r\n"))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	w.state = stateHeadersWritten
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateHeadersWritten {
		return 0, ERROR_INVALID_WRITE_ORDER
	}

	n, err := w.writer.Write(p)
	if err != nil {
		return n, err
	}

	w.state = stateBodyWritten
	w.state = stateBodyWritten
	return n, nil
}
