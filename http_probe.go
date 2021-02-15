package server

func NewHTTPProbe(path, password string, handlers HTTPProbeHandlers) HTTPProbe {
	probe := HTTPProbe{Handlers: handlers}
	probe.Path = path
	probe.Password = password
	return probe
}

type HTTPProbe struct {
	Handlers HTTPProbeHandlers
	HTTPPath
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
