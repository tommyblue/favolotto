package favolotto

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	host string
	port int
}

func NewHTTPServer(host string, port int) *HTTPServer {
	return &HTTPServer{
		host: host,
		port: port,
	}
}

func (s *HTTPServer) Run(ctx context.Context) {
	log.Printf("Starting HTTP server on %s:%d", s.host, s.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})
	mux.Handle("/api/v1", handleSomething())

	httpServer := &http.Server{
		Addr:           net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port)),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}
}

func handleSomething() http.Handler {
	// thing := prepareThing()
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// use thing to handle request
			// logger.Info(r.Context(), "msg", "handleSomething")
			fmt.Fprintf(w, "API v1")
		},
	)
}
