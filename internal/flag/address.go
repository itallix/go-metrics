package flag

import (
	"strconv"
	"strings"
)

const (
	DefaultHost = "localhost"
	DefaultPort = 8080
)

// RunAddress describes Host:Port combination to run on.
type RunAddress struct {
	Host string
	Port int
}

// NewRunAddress constructs new instance of RunAddress with default values.
func NewRunAddress() *RunAddress {
	return &RunAddress{
		Host: DefaultHost,
		Port: DefaultPort,
	}
}

// String returns string representation, example "localhost:8080".
func (a *RunAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

// Set parses string passed in a paramter to a combination of Host and Port values.
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
