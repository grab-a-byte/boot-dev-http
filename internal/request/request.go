package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"dev.grab-a-byte.network/internal/headers"
)

var ERROR_INVALID_REQUEST_LINE = errors.New("invalid request line")
var ERROR_INVLID_HTTP_VERSION = errors.New("invalid http version")
var ERROR_INVALID_HTTP_METHOD = errors.New("invalid http method")
var ERROR_REQUEST_IN_ERROR_STATE = errors.New("request in error state")

type requestStatus string

const (
	StatusInit         requestStatus = "init"
	StatusParseHeaders requestStatus = "parse_headers"
	StatusParseBody    requestStatus = "parse_body"
	StatusDone         requestStatus = "done"
	StatusError        requestStatus = "error"
)

var SEPERATOR = []byte("\r\n")

func newRequest() *Request {
	return &Request{
		status:  StatusInit,
		Headers: headers.NewHeaders(),
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	status      requestStatus
	previousContentLength int
}

func (r *Request) String() string {
	builder := strings.Builder{}
	builder.WriteString("Request line:\n")
	fmt.Fprintf(&builder, "- Method: %s\n", r.RequestLine.Method)
	fmt.Fprintf(&builder, "- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Fprintf(&builder, "- Version: %s\n", r.RequestLine.HttpVersion)
	builder.WriteString("Headers:\n")
	for key, value := range r.Headers {
		value := fmt.Sprintf("- %s: %s\n", key, value)
		builder.WriteString(value)
	}

	builder.WriteString("Body:\n")
	builder.Write(r.Body)
	return builder.String()
}

func (r *Request) done() bool {
	return r.status == StatusDone || r.status == StatusError
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.status {
		case StatusError:
			return 0, fmt.Errorf("Request in error state")
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

			r.status = StatusParseHeaders
		case StatusParseHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			read += n

			if done {
				r.status = StatusParseBody
			}

		case StatusParseBody:
			value, ok := r.Headers.Get("content-length")
			if !ok {
				r.status = StatusDone
				return 0, nil
			}

			if len(data) == r.previousContentLength {
				return 0, fmt.Errorf("Not enough data sent")
			}
			r.previousContentLength = len(data)

			length, err := strconv.Atoi(value)
			if err != nil {
				return 0, err
			}

			if len(data) == 0 {
				return 0, fmt.Errorf("Not enough data sent as part of body")
			}
			if len(data) < length {
				break outer
			}
			r.Body = append(r.Body, data[0:length]...)

			read += len(data)
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
		if err != nil && err != io.EOF {
			return nil, err
		}

		bufLen += n

		readN, err := req.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return req, nil
}

func parseRequestLine(input []byte) (*RequestLine, int, error) {
	before, _, ok := bytes.Cut(input, SEPERATOR)
	if !ok {
		return nil, 0, nil
	}

	read := len(before) + len(SEPERATOR)

	requestLineParts := bytes.Split(before, []byte{' '})

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
