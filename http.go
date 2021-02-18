package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/usvc/go-server/middleware"
)

func NewHTTP(opts HTTPOptions, mux *http.ServeMux) *HTTP {
	addr := opts.Addr.String()
	errorLogger := log.New(loggerFromExternalLogger{Print: opts.Loggers.ServerEvent}, "", 0)

	mux.HandleFunc(opts.LivenessProbe.Path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\"ok\""))
	})

	mux.HandleFunc(opts.ReadinessProbe.Path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if errs := opts.ReadinessProbe.Handlers.Do(); errs != nil {
			errsAsJSON, marshalError := json.Marshal(errs)
			if marshalError != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("\"%s\"", marshalError.Error())))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errsAsJSON)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("\"ok\""))
	})

	mux.HandleFunc(opts.Metrics.Path, func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	mux.HandleFunc(opts.Version.Path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("%s", opts.Version.Value)))
	})

	handler := http.Handler(mux)

	middlewares := middleware.Middlewares{}
	if !opts.Disable.CORS {
		middlewares = append(middlewares, middleware.NewCORS(opts.CORS))
	}
	if !opts.Disable.RequestLogger {
		middlewares = append(middlewares, middleware.NewRequestLogger(middleware.RequestLoggerConfiguration{Log: opts.Loggers.Request}))
	}
	if !opts.Disable.RequestIdentifier {
		middlewares = append(middlewares, middleware.NewRequestIdentifier(middleware.RequestIdentifierConfiguration{}))
	}
	for i := 0; i < len(middlewares); i++ {
		apply := middlewares[i]
		handler = apply(handler)
	}
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
	events := make(chan error)
	sigs := make(chan os.Signal, 1)
	defer func() {
		close(events)
		close(sigs)
	}()
	h.Started = &sync.WaitGroup{}
	h.Started.Add(1)
	go func() {
		h.Server.ErrorLog.Printf("starting server on '%s'...", h.Options.Addr.String())
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

func (h *HTTP) handleEvents(events <-chan error) {
	for {
		event := <-events
		eventMessage := event.Error()
		switch true {
		case strings.Contains(eventMessage, "http: Server closed"):
			h.Server.ErrorLog.Printf("server was closed")
			h.Started.Done()
		case strings.Contains(eventMessage, "received signal: "):
			h.Server.ErrorLog.Printf("server %s", eventMessage)
			h.handleShutdown(event)
			h.Started.Done()
		case strings.Contains(eventMessage, "bind: address already in use"):
			h.Server.ErrorLog.Printf("failed to start server: '%s' is already in use", h.Options.Addr.String())
			h.handleShutdown(event)
			h.Started.Done()
		default:
			h.Server.ErrorLog.Printf("unknown event: %s", event)
		}
	}
}

func (h *HTTP) handleShutdown(event error) []error {
	errors := []error{}
	if h.Options.ShutdownHandlers != nil {
		h.Server.ErrorLog.Printf("running %v shutdown handlers...", len(h.Options.ShutdownHandlers))
		for index, shutdownHandler := range h.Options.ShutdownHandlers {
			if err := shutdownHandler(event); err != nil {
				h.Server.ErrorLog.Printf("shutdown handler %v failed with: %s", index, err)
				errors = append(errors, err)
				continue
			}
			h.Server.ErrorLog.Printf("shutdown handler %v succeeded", index)
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}
