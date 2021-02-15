package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func NewHTTP(opts HTTPOptions, mux *http.ServeMux) *HTTP {
	addr := opts.Addr.String()
	errorLogger := log.New(httplog{Print: opts.Loggers.Error}, "", 0)
	mux.HandleFunc(opts.LivenessProbe.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("\"ok\""))
	})
	mux.HandleFunc(opts.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte("\"todo\""))
	})
	mux.HandleFunc(opts.ReadinessProbe.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte("\"ok\""))
	})
	mux.HandleFunc(opts.Version.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s", opts.Version.Value)))
	})
	handler := http.Handler(mux)
	s := HTTP{
		Options: &opts,
		Server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			ErrorLog:          errorLogger,
			IdleTimeout:       opts.Timeouts.Idle,
			ReadTimeout:       opts.Timeouts.Read,
			ReadHeaderTimeout: opts.Timeouts.ReadHeader,
			WriteTimeout:      opts.Timeouts.Write,
		},
	}
	return &s
}

type HTTP struct {
	Started *sync.WaitGroup
	Options *HTTPOptions
	Server  *http.Server
}

func (h HTTP) Start() {
	h.Started = &sync.WaitGroup{}
	events := make(chan error)
	sigs := make(chan os.Signal, 1)
	h.Options.ShutdownHandlers = append(h.Options.ShutdownHandlers, func(err error) (e error) {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("%s", r)
			}
		}()
		close(events)
		close(sigs)
		return nil
	})
	h.Started.Add(1)
	go func() {
		events <- h.Server.ListenAndServe()
	}()
	go func() {
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		go func() {
			sig := <-sigs
			events <- fmt.Errorf("received signal: %s", sig.String())
		}()
	}()
	go h.handleEvents(events)
	h.Started.Wait()
}

func (h HTTP) handleEvents(events <-chan error) {
	for {
		event := <-events
		eventMessage := event.Error()
		switch true {
		case strings.Contains(eventMessage, "http: Server closed"):
			h.Server.ErrorLog.Printf("server was closed")
			h.Started.Done()
		case strings.Contains(eventMessage, "received signal: "):
			errorCount := 1
			h.Server.ErrorLog.Printf("\nserver %s", eventMessage)
			if h.Options.ShutdownHandlers != nil {
				h.Server.ErrorLog.Printf("running %v shutdown handlers...", len(h.Options.ShutdownHandlers))
				for index, shutdownHandler := range h.Options.ShutdownHandlers {
					if err := shutdownHandler(event); err != nil {
						h.Server.ErrorLog.Printf("shutdown handler %v failed with: %s", index, err)
						errorCount += 1
						continue
					}
					h.Server.ErrorLog.Printf("shutdown handler %v succeeded", index)
				}
			}
			os.Exit(errorCount)
		case strings.Contains(eventMessage, "bind: address already in use"):
			h.Server.ErrorLog.Printf("failed to start server: '%s' is already in use", h.Options.Addr.String())
			os.Exit(1)
		default:
			h.Server.ErrorLog.Printf("unknown event: %s", event)
			os.Exit(255)
		}
	}
}
