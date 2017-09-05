package net

import (
	"context"
	"time"
)

// SlowOptions is used to delay the network packets.
type SlowOptions struct {
	Mean         time.Duration
	Variance     time.Duration
	Distribution string
}

// DefaultSlowOptions returns a default options.
func DefaultSlowOptions() SlowOptions {
	return SlowOptions{
		Mean:         50 * time.Millisecond,
		Variance:     10 * time.Millisecond,
		Distribution: "normal",
	}
}

// Net is used to control the network.
type Net interface {
	// Drop drops traffic from node.
	Drop(ctx context.Context, node string) error
	// Heal ends all traffic drops and restores network to fast operations.
	Heal(ctx context.Context) error
	// Slow delays the network packets with options.
	Slow(ctx context.Context, opts SlowOptions) error
	// Flaky introduces randomized packet loss.
	Flaky(ctx context.Context) error
	// Fast removes packet loss and delays.
	Fast(ctx context.Context) error
}

// Noop implements Net interface but does nothing.
type Noop struct {
	Net
}

// Drop drops traffic from node.
func (Noop) Drop(ctx context.Context, node string) error { return nil }

// Heal ends all traffic drops and restores network to fast operations.
func (Noop) Heal(ctx context.Context) error { return nil }

// Slow delays the network packets with opetions.
func (Noop) Slow(ctx context.Context, opts SlowOptions) error { return nil }

// Flaky introduces randomized packet loss.
func (Noop) Flaky(ctx context.Context) error { return nil }

// Fast removes packet loss and delays.
func (Noop) Fast(ctx context.Context) error { return nil }
