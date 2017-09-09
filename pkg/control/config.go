package control

import (
	"errors"
	"time"
)

// Config is the configuration for the controller.
type Config struct {
	// NodePort is used to communicate with the node server.
	NodePort int
	// DB is the name which we want to run, you must register the db in the node before.
	DB string
	// RequestCount controls how many requests a client sends to the db
	RequestCount int
	// RunTime controls how long the controller takes.
	RunTime time.Duration
}

func (c *Config) adjust() error {
	if len(c.DB) == 0 {
		return errors.New("empty database")
	}

	if c.RequestCount == 0 {
		c.RequestCount = 10000
	}

	if c.RunTime == 0 {
		c.RunTime = 10 * time.Minute
	}

	return nil
}
