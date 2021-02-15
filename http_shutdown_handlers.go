package server

type HTTPShutdownHandlers []HTTPShutdownHandler
type HTTPShutdownHandler func(error) error
