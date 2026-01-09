package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       parserState
}

var ERROR_BAD_START_LINE = fmt.Errorf("bad request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http version")
var ERROR_UNSUPPORTED_HTTP_METHOD = fmt.Errorf("unsupported http method")
var SEPARATOR = []byte("\r\n")

type parserState string

const (
	StateInit parserState = "init"
	StateDone parserState = "done"
)


func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	requestLine := b[:idx]
	read := idx+len(SEPARATOR)

	parts := bytes.Split(requestLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_BAD_START_LINE
	}

	var rl RequestLine

	rl.Method = string(parts[0])
	validMethods := []string{"GET", "POST", "PATCH", "PUT"}
	if !slices.Contains(validMethods, rl.Method) {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_METHOD
	}

	rl.RequestTarget = string(parts[1])

	if string(parts[2]) != "HTTP/1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}
	rl.HttpVersion = strings.Split(string(parts[2]), "/")[1]

	return &rl, read, nil

}

func (r *Request) parse(data []byte) (int, error){
	read := 0
	outer:
	switch r.state {
		case StateInit: 
		rl, n, err := parseRequestLine(data[read:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break outer
		}
		r.RequestLine = *rl
		read += n

		r.state = StateDone
	case StateDone:
		break outer
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func NewRequest() *Request {
	return &Request{
		state: StateInit,
	}
}


func RequestFromReader(reader io.Reader) (*Request, error) {

	request := NewRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			return nil, errors.Join(fmt.Errorf("unable to read"), err)
		}

		bufLen += n
		readN, err := request.parse(buf[:bufLen])

		if err != nil && err != io.EOF {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
