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

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("content-length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/html")

	return h
}

const (
	start = iota
	statusLineWritten
	headersWritten
	done
)

type Writer struct {
	writer io.Writer
	status int
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		status: start}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.status > start {
		return fmt.Errorf("Status line already written")
	}
	w.writer.Write([]byte("HTTP/1.1 "))
	fmt.Fprintf(w.writer, "%d ", statusCode)
	switch statusCode {
	case STATUS_OK:
		w.writer.Write([]byte("OK"))
	case STATUS_BAD_REQUEST:
		w.writer.Write([]byte("Bad Request"))
	case STATUS_INTERNAL_SERVER_ERROR:
		w.writer.Write([]byte("Internal Server Error"))
	}

	w.writer.Write([]byte("\r\n"))
	w.status = statusLineWritten
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.status < statusLineWritten {
		return fmt.Errorf("Need to write status line first")
	}
	// if w.status > statusLineWritten {
	// 	return fmt.Errorf("Headers already written")
	// }
	for k, v := range headers {
		line := fmt.Sprintf("%s: %s\r\n", k, v)
		n, err := w.writer.Write([]byte(line))
		if err != nil {
			return err
		}

		if n == 0 {
			errMsg := fmt.Sprintf("Unable to write header %s", line)
			return fmt.Errorf(errMsg)
		}
	}
	w.writer.Write([]byte("\r\n"))
	w.status = headersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.status < headersWritten {
		return 0, fmt.Errorf("Not ready to write body yet, write headers first")
	}
	if w.status == done {
		return 0, fmt.Errorf("Already written body")
	}
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}

	w.status = done
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	count := fmt.Sprintf("%x\r\n", len(p))
	n, err := w.writer.Write([]byte(count))
	if err != nil {
		return 0, err
	}
	c, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	r, err := w.writer.Write([]byte("\r\n"))
	if err != nil {
		return 0, err
	}

	return n + c + r, nil
}

func (w *Writer) AddCrLf(){
	w.writer.Write([]byte("\r\n"))
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.writer.Write([]byte("0\r\n"))
	return n, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	return w.WriteHeaders(h)
}
