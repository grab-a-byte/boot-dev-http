package main

import (
	"fmt"
	"log"
	"net"

	"dev.grab-a-byte.network/internal/request"
)

func main() {

	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening on port 42069, %q", err)
	}
	log.Println("listening on port 42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}
		log.Println("Connection has been made")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(req.String())
	}
}
