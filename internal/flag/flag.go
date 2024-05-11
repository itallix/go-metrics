package flag

import (
	"errors"
	"strconv"
	"strings"
)

type RunAddress struct {
	Host string
	Port int
}

func NewRunAddress() *RunAddress {
	return &RunAddress{
		Host: "localhost",
		Port: 8080,
	}
}

func (a *RunAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *RunAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}
