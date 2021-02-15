package server

import (
	"log"
	"net/http"
	"time"

	"github.com/usvc/go-server/middleware"
)

func NewHTTPOptions() HTTPOptions {
	return HTTPOptions{
		Addr: HTTPAddr{
			Address: "0.0.0.0",
			Port:    8000,
		},
		CORS: middleware.CORSConfiguration{
			AllowHeaders:      []string{},
			AllowMethods:      []string{http.MethodGet, http.MethodOptions, http.MethodPost},
			AllowOrigins:      []string{"127.0.0.1"},
			AllowCredentials:  false,
			EnablePassthrough: false,
			MaxAge:            30 * time.Minute,
		},
		Disable: HTTPDisables{
			CORS:              false,
			RequestIdentifier: false,
			RequestLogger:     false,
		},
		LivenessProbe: NewHTTPProbe("/healthz", "", nil),
		Loggers: HTTPLoggers{
			ServerEvent: log.Print,
			Request:     log.Print,
		},
		Metrics: HTTPPath{
			Path: "/metrics",
		},
		ReadinessProbe: NewHTTPProbe("/readyz", "", nil),
		Timeouts: HTTPTimeouts{
			Idle:       30 * time.Second,
			Read:       3 * time.Second,
			ReadHeader: 3 * time.Second,
			Write:      10 * time.Second,
		},
		Version: HTTPVersion{
			Path:  "/version",
			Value: "development",
		},
	}
}

type HTTPOptions struct {
	Addr             HTTPAddr                     `json:"addr" yaml:"addr"`
	CORS             middleware.CORSConfiguration `json:"cors" yaml:"cors"`
	Disable          HTTPDisables                 `json:"enable" yaml:"enable"`
	LivenessProbe    HTTPProbe                    `json:"livenessProbe" yaml:"livenessProbe"`
	Metrics          HTTPPath                     `json:"metrics" yaml:"metrics"`
	ReadinessProbe   HTTPProbe                    `json:"readinessProbe" yaml:"readinessProbe"`
	Timeouts         HTTPTimeouts                 `json:"timeouts" yaml:"timeouts"`
	Version          HTTPVersion                  `json:"version" yaml:"version"`
	Middlewares      middleware.Middlewares
	ShutdownHandlers HTTPShutdownHandlers
	Loggers          HTTPLoggers
}

type HTTPDisables struct {
	CORS              bool `json:"cors" yaml:"cors"`
	RequestIdentifier bool `json:"requestIdentifier" yaml:"requestIdentifier"`
	RequestLogger     bool `json:"requestLogger" yaml:"requestLogger"`
}

type HTTPTimeouts struct {
	Idle       time.Duration `json:"idle" yaml:"idle"`
	Read       time.Duration `json:"read" yaml:"read"`
	Write      time.Duration `json:"write" yaml:"write"`
	ReadHeader time.Duration `json:"readHeader" yaml:"readHeader"`
}

type HTTPVersion struct {
	Path     string `json:"path" yaml:"path"`
	Password string `json:"password" yaml:"password"`
	Value    string `json:"value" yaml:"value"`
}
