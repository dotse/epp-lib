package epplib

import "fmt"

// Logger hold all logging functions needed.
type Logger interface {
	Debugf(string, ...interface{})
	Errorf(string, ...interface{})
	Infof(string, ...interface{})
}

// DummyLogger for the logger interface used for testing.
type DummyLogger struct{}

// Errorf for DummyLogger fmt.Printf the information.
func (*DummyLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...) //nolint:forbidigo // dummy logger
}

// Infof for DummyLogger fmt.Printf the information.
func (*DummyLogger) Infof(format string, args ...interface{}) {
	fmt.Printf(format, args...) //nolint:forbidigo // dummy logger
}

// Debugf for DummyLogger fmt.Printf the information.
func (*DummyLogger) Debugf(format string, args ...interface{}) {
	fmt.Printf(format, args...) //nolint:forbidigo // dummy logger
}
