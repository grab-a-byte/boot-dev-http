package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dev.grab-a-byte.network/internal/request"
	"dev.grab-a-byte.network/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(
		port,
		func(w io.Writer, req *request.Request) *server.HandlerError {
			path := req.RequestLine.RequestTarget
			if strings.Contains(path, "/yourproblem") {
				return &server.HandlerError{
					StatusCode:   400,
					ErrorMessage: "Your problem is not my problem\n",
				}
			}
			if strings.Contains(path, "/myproblem") {
				return &server.HandlerError{
					StatusCode:   500,
					ErrorMessage: "Woopsie, my bad\n",
				}
			}

			w.Write([]byte("All good, frfr\n"))
			return nil
		},
	)
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
