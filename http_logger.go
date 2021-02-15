package server

import "github.com/usvc/go-server/types"

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
