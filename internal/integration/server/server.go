// Package server implements a helper server for WebRTC integration testing
package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pions/webrtc/internal/integration"
)

type Server struct {
	Port       string
	Dir        string
	resultChan chan (error)
	server     *http.Server
	sigIn      chan []byte
	sigOut     chan []byte
	res        chan integration.Result
}

const jslib = `
function PostResult(context, message) {
	let data = {
		context: context,
		message: message
	}
	console.log(JSON.stringify(data))
	fetch("/result", {
        method: "POST",
        cache: "no-cache",
        headers: { "Content-Type": "application/json; charset=utf-8" },
        body: JSON.stringify(data),
    })
}

function Signal(description) {
	fetch("/signal", {
        method: "POST",
        cache: "no-cache",
        headers: { "Content-Type": "application/json; charset=utf-8" },
        body: JSON.stringify(description),
    })
}

function OnSignal(callback) {
	fetch("/onsignal", {
        cache: "no-cache",
    })
	.then(response => response.json())
	.then(response => callback(response))
}
`

// New creates a new integration testing server
func New(options ...Option) *Server {
	res := &Server{
		Port: "9090",
		Dir:  "static",
	}

	for _, option := range options {
		option(res)
	}

	return res
}

// Option allows overwriting the default options
type Option func(*Server)

// WithPort allows overwriting the default port (9090)
func WithPort(port string) Option {
	return func(c *Server) {
		c.Port = port
	}
}

// Spawn spawns the server
func (s *Server) Spawn() error {
	mux := http.NewServeMux()

	strip := "/" + s.Dir + "/"
	mux.Handle(strip, http.StripPrefix(strip, http.FileServer(http.Dir(s.Dir))))
	mux.HandleFunc("/integration", func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, jslib)
		if err != nil {
			fmt.Println("Failed to write lib: %v", err)
		}
	})

	mux.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var res integration.Result
		err := json.NewDecoder(r.Body).Decode(&res)

		if err != nil {
			fmt.Printf("failed to unmarshal result: %v\n", err)
			return
		}

		s.res <- res
	})

	mux.HandleFunc("/signal", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		data, _ := ioutil.ReadAll(r.Body)

		s.sigIn <- data
	})

	mux.HandleFunc("/onsignal", func(w http.ResponseWriter, r *http.Request) {
		data := <-s.sigOut

		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write(data)
		if err != nil {
			fmt.Println("Failed to write signal: %v", err)
		}
	})

	server := &http.Server{
		Addr:           ":" + s.Port,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.sigIn = make(chan []byte)
	s.sigOut = make(chan []byte)
	s.res = make(chan integration.Result)

	go server.ListenAndServe()

	return nil
}

func (s *Server) Signal(d []byte) {
	s.sigOut <- d
}

func (s *Server) OnSignal() []byte {
	return <-s.sigIn
}

func (s *Server) Listen() integration.Result {
	return <-s.res
}

// Close closes the server
func (s *Server) Close() error {
	if s.server == nil {
		return nil
	}
	err := s.server.Close()
	if err != nil {
		return err
	}
	s.server = nil
	return nil
}
