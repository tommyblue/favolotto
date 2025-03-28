package favolotto

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	host  string
	port  int
	store *Store
}

func NewHTTPServer(host string, port int, store *Store) *HTTPServer {
	return &HTTPServer{
		host:  host,
		port:  port,
		store: store,
	}
}

func (s *HTTPServer) Run(ctx context.Context) {
	log.Printf("Starting HTTP server on http://%s:%d", s.host, s.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
		// serve index.html
	})
	mux.Handle("/api/v1/", s.apiMux())

	httpServer := &http.Server{
		Addr:           net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port)),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   300 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	<-ctx.Done()

	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error shutting down server: %s\n", err)
	}

	log.Println("HTTP server stopped")
}

func (s *HTTPServer) apiMux() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /songs", s.listSongs())
	mux.Handle("GET /tags/current", s.currentTag())
	mux.Handle("PUT /song", s.putSong())
	mux.Handle("DELETE /song", s.deleteSong())
	return http.StripPrefix("/api/v1", mux)
}

func (s *HTTPServer) listSongs() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			songs := s.store.getMetadata()
			// use thing to handle request
			// logger.Info(r.Context(), "msg", "handleSomething")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(songs)
		},
	)
}

func (s *HTTPServer) currentTag() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			nfc := s.store.LastNfc()
			resp := struct {
				NfcTag string `json:"nfc_tag"`
			}{
				NfcTag: nfc,
			}
			// use thing to handle request
			// logger.Info(r.Context(), "msg", "handleSomething")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		},
	)
}

func (s *HTTPServer) putSong() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseMultipartForm(150 << 20) // 150 MB
			file, header, err := r.FormFile("song")
			if err != nil {
				http.Error(w, "Error uploading file", http.StatusBadRequest)
				return
			}
			defer file.Close()

			nfcTag := r.FormValue("nfc_tag")
			if nfcTag == "" {
				http.Error(w, "Missing nfc_tag", http.StatusBadRequest)
				return
			}

			if err := s.store.putSong(nfcTag, header.Filename, file); err != nil {
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
		},
	)
}

func (s *HTTPServer) deleteSong() http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// use thing to handle request
			// logger.Info(r.Context(), "msg", "handleSomething")
			type request struct {
				NfcTag string `json:"nfc_tag"`
			}
			var req request

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Error decoding request", http.StatusBadRequest)
				return
			}

			if req.NfcTag == "" {
				http.Error(w, "Missing nfc_tag", http.StatusBadRequest)
				return
			}

			if err := s.store.deleteSong(req.NfcTag); err != nil {
				http.Error(w, fmt.Sprintf("Error deleting song: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
		},
	)
}
