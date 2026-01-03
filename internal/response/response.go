package response

import (
	"fmt"
	"io"
	"strconv"

	"dev.grab-a-byte.network/internal/headers"
)

type StatusCode int

const (
	STATUS_OK                    StatusCode = 200
	STATUS_BAD_REQUEST           StatusCode = 400
	STATUS_INTERNAL_SERVER_ERROR StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) {
	w.Write([]byte("HTTP/1.1 "))
	fmt.Fprintf(w, "%d ", statusCode)
	switch statusCode {
	case STATUS_OK:
		w.Write([]byte("OK"))
	case STATUS_BAD_REQUEST:
		w.Write([]byte("Bad Request"))
	case STATUS_INTERNAL_SERVER_ERROR:
		w.Write([]byte("Internal Server Error"))
	}

	w.Write([]byte("\r\n"))
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		line := fmt.Sprintf("%s: %s\r\n", k, v)
		n, err := w.Write([]byte(line))
		if err != nil {
			return err
		}

		if n == 0 {
			errMsg := fmt.Sprintf("Unable to write header %s", line)
			return fmt.Errorf(errMsg)
		}
	}

	return nil
}
