package server

type HTTPPath struct {
	Path     string `json:"path" yaml:"path"`
	Password string `json:"password" yaml:"password"`
}
