package flag

import (
	"strconv"
	"strings"
)

const (
	DefaultHost = "localhost"
	DefaultPort = 8080
)

type RunAddress struct {
	Host string
	Port int
}

func NewRunAddress() *RunAddress {
	return &RunAddress{
		Host: DefaultHost,
		Port: DefaultPort,
	}
}

func (a *RunAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *RunAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}
