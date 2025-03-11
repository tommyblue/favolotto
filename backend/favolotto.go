package favolotto

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/tommyblue/favolotto/backup/tagreader"
)

type Config struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Favolotto struct {
	config Config
}

func New(config Config) *Favolotto {
	return &Favolotto{
		config: config,
	}

}

func (f *Favolotto) Run(ctx context.Context) error {
	// initialize NFC reader (PN532 via I2C)
	f.nfc(ctx)
	// initialize GPIO (button)
	// initialize GPIO (LED)
	// initialize mp3 player
	// initialize web server
	httpServer := NewHTTPServer(f.config.Host, f.config.Port)
	go httpServer.Run(ctx)

	return nil
}

func (f *Favolotto) nfc(ctx context.Context) {
	// Create an abstraction of the Reader, DeviceConnection string is empty if you want the library to autodetect your reader
	rfidReader := tagreader.NewTagReader("", 19)
	tagChannel := rfidReader.GetTagChannel()

	// Listen for an RFID/NFC tag in a separate goroutine
	go rfidReader.ListenForTags(ctx)

	for {
		select {
		case tagId := <-tagChannel:
			log.Printf("Read tag: %s \n", tagId)
		case <-ctx.Done():
			err := rfidReader.Cleanup()
			if err != nil {
				log.Fatal("Error cleaning up the reader: ", err.Error())
			}
			return
		default:
			log.Printf("%s: Waiting for a tag \n", time.Now().String())
			time.Sleep(time.Millisecond * 300)
		}
	}
}

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
