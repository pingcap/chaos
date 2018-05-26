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
	// Drop runs on the node and drops traffic from srcNode.
	Drop(ctx context.Context, node string, srcNode string) error
	// Heal runs on the node and ends all traffic drops and restores network to fast operations.
	Heal(ctx context.Context, node string) error
	// Slow runs on the node and delays the network packets with options.
	Slow(ctx context.Context, node string, opts SlowOptions) error
	// Flaky runs on the node and introduces randomized packet loss.
	Flaky(ctx context.Context, node string) error
	// Fast runs on the node and removes packet loss and delays.
	Fast(ctx context.Context, node string) error
}

// Noop implements Net interface but does nothing.
type Noop struct {
	Net
}

// Drop drops traffic from node.
func (Noop) Drop(ctx context.Context, node string, srcNode string) error { return nil }

// Heal ends all traffic drops and restores network to fast operations.
func (Noop) Heal(ctx context.Context, node string) error { return nil }

// Slow delays the network packets with opetions.
func (Noop) Slow(ctx context.Context, node string, opts SlowOptions) error { return nil }

// Flaky introduces randomized packet loss.
func (Noop) Flaky(ctx context.Context, node string) error { return nil }

// Fast removes packet loss and delays.
func (Noop) Fast(ctx context.Context, node string) error { return nil }
