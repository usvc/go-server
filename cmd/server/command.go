package main

import (
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/usvc/go-server"
)

func GetCommand() *cobra.Command {
	command := cobra.Command{
		Use:  "server",
		Long: "Example code for the github.com/usvc/go-server module",
		Run: func(cmd *cobra.Command, _ []string) {
			options := server.NewHTTPOptions()
			options.CORS.AllowHeaders = []string{"X-Auth"}
			options.LivenessProbe.Handlers = []server.HTTPProbeHandler{
				func() error {
					<-time.After(1 * time.Second)
					RequestLogger("example liveness probe 1")
					return nil
				},
				func() error {
					<-time.After(1 * time.Second)
					RequestLogger("example liveness probe 2")
					return nil
				},
			}
			options.Loggers.ServerEvent = ServerEventLogger
			options.Loggers.Request = RequestLogger
			options.ReadinessProbe.Handlers = []server.HTTPProbeHandler{
				func() error {
					<-time.After(1 * time.Second)
					RequestLogger("example readiness probe 1")
					return nil
				},
				func() error {
					<-time.After(1 * time.Second)
					RequestLogger("example readiness probe 2")
					return nil
				},
			}
			mux := http.NewServeMux()
			s := server.NewHTTP(options, mux)
			s.Start()
		},
	}
	return &command
}
