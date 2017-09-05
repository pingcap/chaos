package net

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// IPTables implements Net interface to simulate the network.
type IPTables struct {
	Net
}

// Drop drops traffic from node.
func (IPTables) Drop(ctx context.Context, node string) error {
	return exec.CommandContext(ctx, "iptables", "-A", "INPUT", "-s", HostIP(node), "-j", "DROP", "-w").Run()
}

// Heal ends all traffic drops and restores network to fast operations.
func (IPTables) Heal(ctx context.Context) error {
	if err := exec.CommandContext(ctx, "iptables", "-F", "-w").Run(); err != nil {
		return err
	}
	return exec.CommandContext(ctx, "iptables", "-X", "-w").Run()
}

// Slow delays the network packets with opetions.
func (IPTables) Slow(ctx context.Context, opts SlowOptions) error {
	mean := fmt.Sprintf("%dms", opts.Mean.Nanoseconds()/int64(time.Millisecond))
	variance := fmt.Sprintf("%dms", opts.Variance.Nanoseconds()/int64(time.Millisecond))
	return exec.CommandContext(ctx, "/sbin/tc", "qdisc", "add", "dev", "eth0", "root", "netem", "delay",
		mean, variance, "distribution", opts.Distribution).Run()
}

// Flaky introduces randomized packet loss.
func (IPTables) Flaky(ctx context.Context) error {
	return exec.CommandContext(ctx, "/sbin/tc", "qdisc", "add", "dev", "eth0", "root", "netem", "loss",
		"20%", "75%").Run()
}

// Fast removes packet loss and delays.
func (IPTables) Fast(ctx context.Context) error {
	output, err := exec.CommandContext(ctx, "/sbin/tc", "qdisc", "del", "dev", "eth0", "root").CombinedOutput()
	if err != nil && strings.Contains(string(output), "RTNETLINK answers: No such file or directory") {
		err = nil
	}
	return err
}
