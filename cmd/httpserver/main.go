package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dev.grab-a-byte.network/internal/headers"
	"dev.grab-a-byte.network/internal/request"
	"dev.grab-a-byte.network/internal/response"
	"dev.grab-a-byte.network/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(
		port,
		func(w *response.Writer, req *request.Request) {
			path := req.RequestLine.RequestTarget
			defaultHeaders := response.GetDefaultHeaders(0)
			if strings.Contains(path, "/yourproblem") {
				defaultHeaders.Set("Content-Length", fmt.Sprintf("%d", len(badRequestHtml)))
				w.WriteStatusLine(400)
				w.WriteHeaders(defaultHeaders)
				w.WriteBody([]byte(badRequestHtml))
				return
			}
			if strings.Contains(path, "/myproblem") {
				defaultHeaders.Set("Content-Length", fmt.Sprintf("%d", len(internalServerErrorHtml)))
				w.WriteStatusLine(500)
				w.WriteHeaders(defaultHeaders)
				w.WriteBody([]byte(internalServerErrorHtml))
				return
			}

			if strings.Contains(path, "/video") {
				bytes, err := os.ReadFile("../.././assets/vim.mp4")
				if err != nil {
					slog.Info("%s", err)
					return
				}
				defaultHeaders.Set("Content-Length", fmt.Sprintf("%d",len(bytes)))
				defaultHeaders.Set("Content-Type", "video/mp4")
				w.WriteStatusLine(200)
				w.WriteHeaders(defaultHeaders)
				w.WriteBody(bytes)
				return
			}

			if after, ok := strings.CutPrefix(path, "/httpbin/"); ok {
				w.WriteStatusLine(response.STATUS_OK)
				defaultHeaders.Remove("Content-Length")
				defaultHeaders.Set("Transfer-Encoding", "chunked")
				defaultHeaders.Set("Trailer", "X-Content-SHA256, X-Content-Length")
				w.WriteHeaders(defaultHeaders)
				proxied := "https://httpbin.org/" + after
				res, err := http.Get(proxied)
				if err != nil {
					return
				}
				buf := make([]byte, 1024)
				total := strings.Builder{}
				for {
					n, err := res.Body.Read(buf)
					if err != nil {
						break
					}
					total.Write(buf[:n])
					w.WriteChunkedBody(buf[:n])
				}
				w.WriteChunkedBodyDone()
				content := total.String()
				contentLen := len(content)
				contentSha := sha256.Sum256([]byte(content))

				trailers := headers.NewHeaders()
				trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", contentSha))
				trailers.Set("X-Content-Length", fmt.Sprintf("%d", contentLen))
				w.WriteTrailers(trailers)
				return
			}

			defaultHeaders.Set("Content-Length", fmt.Sprintf("%d", len(okHtml)))
			w.WriteStatusLine(500)
			w.WriteHeaders(defaultHeaders)
			w.WriteBody([]byte(okHtml))
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

const badRequestHtml = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalServerErrorHtml = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const okHtml = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
