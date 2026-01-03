package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"dev.grab-a-byte.network/internal/request"
	"dev.grab-a-byte.network/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	ser := &Server{
		listener: listener,
		closed:   atomic.Bool{},
		handler:  handler,
	}

	ser.closed.Store(false)

	go ser.listen()
	return ser, nil
}

func (s *Server) Close() error {
	if s.closed.Load() {
		return fmt.Errorf("Server already closed")
	}
	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			panic("Unable to accept connection")
		}
		if s.closed.Load() {
			fmt.Println("Server closed, ending accepting connections")
			break
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		conn.Write([]byte("Failed to read request"))
		conn.Close()
		return
	}

	buf := &bytes.Buffer{}
	handleErr := s.handler(buf, req)
	if handleErr != nil {
		conn.Write([]byte(handleErr.Error()))
		conn.Close()
		return
	}

	headers := response.GetDefaultHeaders(buf.Len())
	response.WriteStatusLine(conn, response.STATUS_OK)
	response.WriteHeaders(conn, headers)
	conn.Write([]byte("\r\n"))
	conn.Write(buf.Bytes())
	err = conn.Close()
	if err != nil {
		panic("Failure closing connection")
	}
}

type HandlerError struct {
	StatusCode   int
	ErrorMessage string
}

func (he *HandlerError) Error() string {
	return fmt.Sprintf("Error from handler. %d: %s", he.StatusCode, he.ErrorMessage)
}

type Handler func(w io.Writer, req *request.Request) *HandlerError
