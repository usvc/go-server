package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/usvc/go-server/middleware"
	"github.com/usvc/go-server/types"
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
		Disable: HTTPDisable{
			CORS:              false,
			RequestIdentifier: false,
			RequestLogger:     false,
		},
		LivenessProbe: HTTPProbe{
			Path:     "/healthz",
			Password: "",
			Handlers: nil,
		},
		Loggers: HTTPLoggers{
			ServerEvent: log.Print,
			Request:     log.Print,
		},
		Metrics: HTTPPath{
			Path: "/metrics",
		},
		ReadinessProbe: HTTPProbe{
			Path:     "/healthz",
			Password: "",
			Handlers: nil,
		},
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
	Disable          HTTPDisable                  `json:"enable" yaml:"enable"`
	Limit            HTTPLimit                    `json:"limit" yaml:"limit"`
	LivenessProbe    HTTPProbe                    `json:"livenessProbe" yaml:"livenessProbe"`
	Metrics          HTTPPath                     `json:"metrics" yaml:"metrics"`
	ReadinessProbe   HTTPProbe                    `json:"readinessProbe" yaml:"readinessProbe"`
	Timeouts         HTTPTimeouts                 `json:"timeouts" yaml:"timeouts"`
	Version          HTTPVersion                  `json:"version" yaml:"version"`
	Middlewares      middleware.Middlewares
	ShutdownHandlers HTTPShutdownHandlers
	Loggers          HTTPLoggers
}

type HTTPAddr struct {
	Address string `json:"address" yaml:"address"`
	Port    uint   `json:"port" yaml:"port"`
}

func (httpaddr HTTPAddr) String() string {
	return fmt.Sprintf("%s:%v", httpaddr.Address, httpaddr.Port)
}

type HTTPDisable struct {
	CORS              bool `json:"cors" yaml:"cors"`
	LivenessProbe     bool `json:"livenessProbe" yaml:"livenessProbe"`
	Metrics           bool `json:"metrics" yaml:"metrics"`
	ReadinessProbe    bool `json:"readinessProbe" yaml:"readinessProbe"`
	RequestIdentifier bool `json:"requestIdentifier" yaml:"requestIdentifier"`
	RequestLogger     bool `json:"requestLogger" yaml:"requestLogger"`
	Version           bool `json:"version" yaml:"version"`
}

type HTTPLimit struct {
	HeaderBytes int `json:"headerBytes" yaml:"headerBytes"`
}

type HTTPLoggers struct {
	ServerEvent types.Logger
	Request     types.Logger
}

type loggerFromExternalLogger struct {
	Print types.Logger
}

func (lfel loggerFromExternalLogger) Write(what []byte) (int, error) {
	lfel.Print(string(what))
	return len(what), nil
}

type HTTPPath struct {
	Password string `json:"password" yaml:"password"`
	Path     string `json:"path" yaml:"path"`
}

type HTTPProbe struct {
	Handlers HTTPProbeHandlers
	Password string `json:"password" yaml:"password"`
	Path     string `json:"path" yaml:"path"`
}

type HTTPProbeHandler func() error
type HTTPProbeHandlers []HTTPProbeHandler

func (httpph HTTPProbeHandlers) Do() []error {
	errors := []error{}
	for _, handler := range httpph {
		if err := handler(); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return errors
}

type HTTPShutdownHandlers []HTTPShutdownHandler
type HTTPShutdownHandler func(error) error

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
