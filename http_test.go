package server

import (
	"bytes"
	"fmt"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type HTTPTest struct {
	suite.Suite
	latency time.Duration
}

func TestHTTP(t *testing.T) {
	suite.Run(t, &HTTPTest{
		latency: time.Millisecond * 5,
	})
}

func (s HTTPTest) Test_e2e() {
	var serverEvents bytes.Buffer
	o := NewHTTPOptions()
	o.Addr = HTTPAddr{Address: "0.0.0.0", Port: 55555}
	o.Loggers.ServerEvent = func(args ...interface{}) {
		fmt.Fprint(&serverEvents, args...)
	}

	h := http.NewServeMux()
	sv := NewHTTP(o, h)
	s.T().Log("HI")
	go func(after <-chan time.Time) {
		<-after
		sv.Stop()
	}(time.After(s.latency))
	sv.Start()
	serverEventsLog := serverEvents.String()
	s.Contains(serverEventsLog, "starting server on")
	s.Contains(serverEventsLog, "server was closed")

	serverEvents.Reset()

	h = http.NewServeMux()
	sv = NewHTTP(o, h)
	go func(after <-chan time.Time) {
		<-after
		sv.signals <- syscall.SIGTERM
	}(time.After(s.latency))
	sv.Start()
	serverEventsLog = serverEvents.String()
	s.Contains(serverEventsLog, "server received signal: terminated")

	serverEvents.Reset()

	h = http.NewServeMux()
	sv = NewHTTP(o, h)
	go func(after <-chan time.Time) {
		<-after
		sv.signals <- syscall.SIGINT
	}(time.After(s.latency))
	sv.Start()
	serverEventsLog = serverEvents.String()
	s.Contains(serverEventsLog, "server received signal: interrupt")

	serverEvents.Reset()

	var serverEvents2 bytes.Buffer
	o2 := NewHTTPOptions()
	o2.Addr = HTTPAddr{Address: "0.0.0.0", Port: 55555}
	o2.Loggers.ServerEvent = func(args ...interface{}) {
		fmt.Fprint(&serverEvents2, args...)
	}
	h = http.NewServeMux()
	sv = NewHTTP(o, h)
	h2 := http.NewServeMux()
	sv2 := NewHTTP(o2, h2)
	go func(after <-chan time.Time) {
		<-after
		sv2.Start()
		go func(after2 <-chan time.Time) {
			<-after2
			sv.Stop()
		}(time.After(s.latency))
	}(time.After(s.latency))
	sv.Start()
	serverEventsLog = serverEvents.String()
	serverEvents2Log := serverEvents2.String()
	s.Contains(serverEvents2Log, "'0.0.0.0:55555' is already in use")
	s.Contains(serverEventsLog, "server was closed")
}
