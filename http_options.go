package server

import (
	"fmt"
	"log"
	"time"
)

func NewHTTPOptions() HTTPOptions {
	return HTTPOptions{
		Addr: HTTPAddr{
			Address: "0.0.0.0",
			Port:    8000,
		},
		LivenessProbe: HTTPPath{
			Path: "/healthz",
		},
		Loggers: HTTPLoggers{
			Error: log.Print,
		},
		Metrics: HTTPPath{
			Path: "/metrics",
		},
		ReadinessProbe: HTTPPath{
			Path: "/readyz",
		},
		Timeouts: HTTPTimeouts{
			Idle:       30 * time.Second,
			Read:       3 * time.Second,
			ReadHeader: 3 * time.Second,
			Write:      10 * time.Second,
		},
		Version: HTTPVersion{
			Value: "development",
		},
	}
}

type HTTPOptions struct {
	Addr             HTTPAddr     `json:"addr" yaml:"addr"`
	LivenessProbe    HTTPPath     `json:"livenessProbe" yaml:"livenessProbe"`
	Metrics          HTTPPath     `json:"metrics" yaml:"metrics"`
	ReadinessProbe   HTTPPath     `json:"readinessProbe" yaml:"readinessProbe"`
	Timeouts         HTTPTimeouts `json:"timeouts" yaml:"timeouts"`
	Version          HTTPVersion  `json:"version" yaml:"version"`
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

type HTTPLoggers struct {
	Error   HTTPLogger
	Request HTTPLogger
}

type HTTPPath struct {
	Path     string `json:"path" yaml:"path"`
	Password string `json:"password" yaml:"password"`
}

type HTTPLogger func(args ...interface{})

type httplog struct {
	Print HTTPLogger
}

func (hl httplog) Write(what []byte) (int, error) {
	hl.Print(string(what))
	return len(what), nil
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
