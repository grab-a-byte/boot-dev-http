package request

import (
	"bytes"
	"errors"
	"io"
)

var ERROR_INVALID_REQUEST_LINE = errors.New("invalid request line")
var ERROR_INVLID_HTTP_VERSION = errors.New("invalid http version")
var ERROR_INVALID_HTTP_METHOD = errors.New("invalid http method")
var ERROR_REQUEST_IN_ERROR_STATE = errors.New("request in error state")

type requestStatus string

const (
	StatusInit  requestStatus = "init"
	StatusDone  requestStatus = "done"
	StatusError requestStatus = "error"
)

var SEPERATOR = []byte("\r\n")

func newRequest() *Request {
	return &Request{
		status: StatusInit,
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	status      requestStatus
}

func (r *Request) done() bool {
	return r.status == StatusDone || r.status == StatusError
}

func (r *Request) parse(data []byte) (int, error) {

	read := 0

outer:
	for {
		switch r.status {
		case StatusDone:
			return read, nil
		case StatusInit:
			rl, n, err := parseRequestLine(data)
			if err != nil {
				r.status = StatusError
				return 0, errors.Join(ERROR_REQUEST_IN_ERROR_STATE, err)
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.status = StatusDone
			return read, nil
		}
	}

	return read, nil
}

func RequestFromReader(r io.Reader) (*Request, error) {
	req := newRequest()

	//NOTE: Buffer could overrun (e.g. body or auth token)
	buf := make([]byte, 1024)
	bufLen := 0

	for !req.done() {
		n, err := r.Read(buf[bufLen:])
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := req.parse(buf[:bufLen+n])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return req, nil
}

func parseRequestLine(input []byte) (*RequestLine, int, error) {
	idx := bytes.Index(input, SEPERATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := input[:idx]
	read := len(startLine) + len(SEPERATOR)

	requestLineParts := bytes.Split(startLine, []byte{' '})

	if len(requestLineParts) != 3 {
		return nil, 0, ERROR_INVALID_REQUEST_LINE
	}

	if !allUppercase(requestLineParts[0]) {
		return nil, 0, ERROR_INVALID_HTTP_METHOD
	}

	httpParts := bytes.Split(requestLineParts[2], []byte{'/'})
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_INVLID_HTTP_VERSION
	}

	requestLine := RequestLine{
		HttpVersion:   string(httpParts[1]),
		RequestTarget: string(requestLineParts[1]),
		Method:        string(requestLineParts[0]),
	}

	return &requestLine, read, nil
}

func allUppercase(bytes []byte) bool {
	for _, c := range bytes {
		if c <= 'A' || c >= 'Z' {
			return false
		}
	}

	return true
}
