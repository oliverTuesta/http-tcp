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

var ERROR_INVALID_STATUS_CODE = fmt.Errorf("invalid status code")

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
		case StatusOk,
		StatusBadRequest,
		StatusNotFound,
		StatusInternalServerError:
	default:
		return ERROR_INVALID_STATUS_CODE
	}

	_, err := w.Write([]byte("HTTP/1.1 " + string(statusCode) + "\r\n"))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	var headers = headers.NewHeaders()
	headers.Add("content-length", strconv.Itoa(contentLen))
	headers.Add("connection", "close")
	headers.Add("content-type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(key + ": " + value + "\r\n"))	
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))	
	return err
}

