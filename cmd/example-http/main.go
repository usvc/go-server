package main

import (
	"fmt"
	"time"

	"github.com/usvc/go-server"
)

type logger struct {
}

func (l logger) Print(args ...interface{}) {
	fmt.Print(args...)
}

func main() {
	options := server.NewHTTPOptions()
	options.Loggers.Error = logger{}.Print
	s := server.NewHTTP(options)
	go func(after <-chan time.Time) {
		<-after
		s.Server.Close()
	}(time.After(5 * time.Second))
	s.Start()
}
