package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HTTPTests struct {
	suite.Suite
}

func TestHTTP(t *testing.T) {
	suite.Run(t, &HTTPTests{})
}

func (s HTTPTests) Test_HTTPProbeHandlers() {
	completed := []string{}
	handlers := HTTPProbeHandlers{
		func() error {
			completed = append(completed, "first")
			return nil
		},
		func() error {
			completed = append(completed, "second")
			return nil
		},
	}
	errors := handlers.Do()
	s.Nil(errors)

	completed = []string{}
	handlers = HTTPProbeHandlers{
		func() error {
			return fmt.Errorf("first")
		},
		func() error {
			return fmt.Errorf("second")
		},
	}
	errors = handlers.Do()
	s.Len(errors, 2)
	s.Equal("first", errors[0].Error())
	s.Equal("second", errors[1].Error())
}
