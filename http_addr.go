package server

import "fmt"

type HTTPAddr struct {
	Address string `json:"address" yaml:"address"`
	Port    uint   `json:"port" yaml:"port"`
}

func (httpaddr HTTPAddr) String() string {
	return fmt.Sprintf("%s:%v", httpaddr.Address, httpaddr.Port)
}
