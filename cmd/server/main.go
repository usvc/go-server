package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/usvc/go-server"
)

func ServerEventLogger(args ...interface{}) {
	logrus.Debug(args...)
}

func RequestLogger(args ...interface{}) {
	logrus.Info(args...)
}

func main() {
	logrus.SetLevel(logrus.TraceLevel)
	options := server.NewHTTPOptions()
	options.Loggers.ServerEvent = ServerEventLogger
	options.Loggers.Request = RequestLogger
	options.CORS.AllowHeaders = []string{"X-Auth"}
	options.ReadinessProbe.Handlers = []server.HTTPProbeHandler{
		func() error {
			<-time.After(30 * time.Second)
			return nil
		},
		func() error {
			<-time.After(1 * time.Second)
			return fmt.Errorf("hahha haha")
		},
	}
	mux := http.NewServeMux()
	s := server.NewHTTP(options, mux)
	// go func(after <-chan time.Time) {
	// 	<-after
	// 	s.Stop()
	// }(time.After(5 * time.Second))
	s.Start()
}
