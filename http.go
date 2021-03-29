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

	"github.com/usvc/go-server/handlers"
	"github.com/usvc/go-server/middleware"
)

type FuncHandler interface {
	HandleFunc(string, http.HandlerFunc)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// NewHTTP returns a new HTTP-based server based on the provided options :opts and the
// custom routes handler :mux
func NewHTTP(opts HTTPOptions, mux FuncHandler) *HTTP {
	addr := opts.Addr.String()
	errorLogger := log.New(loggerFromExternalLogger{Print: opts.Loggers.ServerEvent}, "", 0)

	if !opts.Disable.LivenessProbe {
		errorLogger.Print("liveness probe is ENABLED")
		mux.HandleFunc(opts.LivenessProbe.Path, handlers.GetHTTPReadinessProbe(opts.ReadinessProbe.Handlers))
	}

	if !opts.Disable.ReadinessProbe {
		errorLogger.Print("readiness probe is ENABLED")
		mux.HandleFunc(opts.ReadinessProbe.Path, handlers.GetHTTPLivenessProbe(opts.ReadinessProbe.Handlers))
	}

	if !opts.Disable.Metrics {
		errorLogger.Print("metrics is ENABLED")
		mux.HandleFunc(opts.Metrics.Path, handlers.GetHTTPMetrics())
	}

	if !opts.Disable.Version {
		errorLogger.Print("version is ENABLED")
		mux.HandleFunc(opts.Version.Path, handlers.GetHTTPVersion(opts.Version.Value))
	}

	handler := http.Handler(mux)

	middlewares := middleware.Middlewares{}
	if opts.Middlewares != nil && len(opts.Middlewares) > 0 {
		middlewares = append(middlewares, opts.Middlewares...)
	}
	if !opts.Disable.CORS {
		errorLogger.Print("cross-origin resource sharing is ENABLED")
		middlewares = append(middlewares, middleware.NewCORS(opts.CORS))
	}
	if !opts.Disable.RequestLogger {
		errorLogger.Print("request logging is ENABLED")
		middlewares = append(middlewares, middleware.NewRequestLogger(middleware.RequestLoggerConfiguration{Log: opts.Loggers.Request}))
	}
	if !opts.Disable.RequestIdentifier {
		errorLogger.Print("request identification is ENABLED")
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
			MaxHeaderBytes:    opts.Limit.HeaderBytes,
			ReadTimeout:       opts.Timeouts.Read,
			ReadHeaderTimeout: opts.Timeouts.ReadHeader,
			WriteTimeout:      opts.Timeouts.Write,
		},
	}
	return &s
}

// HTTP defines a class for a HTTP-based server
type HTTP struct {
	// Options provides the configuraton for the HTTP server
	Options *HTTPOptions
	// Server points to the raw instance of a http.Server used internally
	Server *http.Server

	// events is a channel for internal communication, avoid subscribing to this since
	// that may cause some events to be missed by internal event handlers
	events chan error
	// signals is a channel to pass system interrupts from process to internal event handlers.
	// to disable this, set the configuration in Options.Disable.SignalHandling
	signals chan os.Signal
}

// Start starts the HTTP-based server
func (h *HTTP) Start() {
	var tasks sync.WaitGroup
	initialise(h)
	defer denitialise(h)
	if !h.Options.Disable.SignalHandling {
		go startSignalsHandler(h)
	}
	go startEventsHandler(h, &tasks)
	tasks.Add(1)
	go startHTTP(h)
	tasks.Wait()
}

// Stop terminates the server process gracefully
func (h *HTTP) Stop() {
	h.events <- h.Server.Close()
}

// denitialise closes the channels that this Server instance uses to communicate events internally
func denitialise(h *HTTP) {
	close(h.events)
	close(h.signals)
}

// initialise initialises the server
func initialise(h *HTTP) {
	h.events = make(chan error)
	h.signals = make(chan os.Signal, 1)
}

// startHTTP starts the server
func startHTTP(h *HTTP) {
	h.Server.ErrorLog.Printf("starting server on '%s'...", h.Options.Addr.String())
	h.events <- h.Server.ListenAndServe()
}

// startSignalsHandler routes system calls like SIGTERM to the server events channel
// for graceful handling
func startSignalsHandler(h *HTTP) {
	signal.Notify(h.signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	if sig := <-h.signals; sig != nil {
		h.events <- fmt.Errorf("received signal: %s", sig.String())
	}
}

// startEventsHandler is an indefinitely looping function meant to be called as a goroutine
// to handle events passed from another sub-routine
func startEventsHandler(h *HTTP, tasks *sync.WaitGroup) {
	for {
		if event := <-h.events; event != nil {
			eventMessage := event.Error()
			fmt.Println(eventMessage)
			switch {
			case strings.Contains(eventMessage, "http: Server closed"):
				h.Server.ErrorLog.Printf("server was closed")
				tasks.Done()
				return
			case strings.Contains(eventMessage, "bind: address already in use"):
				h.Server.ErrorLog.Printf("failed to start server: '%s' is already in use", h.Options.Addr.String())
				handleShutdown(h, event)
				tasks.Done()
				return
			case strings.Contains(eventMessage, "received signal: "):
				h.Server.ErrorLog.Printf("server %s", eventMessage)
				handleShutdown(h, event)
				h.Server.Close()
			default:
				h.Server.ErrorLog.Printf("unknown event: %s", event)
			}
		}
	}
}

// handleShutdown iterates through the shutdown handlers, passing each the provided event :event
// and leaving the handlers to do what they need to before allowing the Server instance to complete
func handleShutdown(h *HTTP, event error) []error {
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
