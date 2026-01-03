package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	c := make(chan string)
	go func() {
		defer close(c)
		defer f.Close()
		defer log.Println("Connection closed")

		line := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)
			if err == io.EOF {
				break
			}

			data = data[:n]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				line += string(data[:i])
				data = data[i+1:]
				c <- line
				line = ""
			}

			line += string(data)
		}
	}()

	return c
}

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error listening on port 42069, %q", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error", "error", err)
		}
		log.Println("Connection has been made")

		for value := range getLinesChannel(conn) {
			fmt.Println(value)
		}
	}
}
