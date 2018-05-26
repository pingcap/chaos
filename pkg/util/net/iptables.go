package net

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/siddontang/chaos/pkg/util/ssh"
)

// IPTables implements Net interface to simulate the network.
type IPTables struct {
	Net
}

// Drop drops traffic from node.
func (IPTables) Drop(ctx context.Context, node string, srcNode string) error {
	return ssh.Exec(ctx, node, "iptables", "-A", "INPUT", "-s", HostIP(srcNode), "-j", "DROP", "-w")
}

// Heal ends all traffic drops and restores network to fast operations.
func (IPTables) Heal(ctx context.Context, node string) error {
	if err := ssh.Exec(ctx, node, "iptables", "-F", "-w"); err != nil {
		return err
	}
	return ssh.Exec(ctx, node, "iptables", "-X", "-w")
}

// Slow delays the network packets with opetions.
func (IPTables) Slow(ctx context.Context, node string, opts SlowOptions) error {
	mean := fmt.Sprintf("%dms", opts.Mean.Nanoseconds()/int64(time.Millisecond))
	variance := fmt.Sprintf("%dms", opts.Variance.Nanoseconds()/int64(time.Millisecond))
	return ssh.Exec(ctx, node, "/sbin/tc", "qdisc", "add", "dev", "eth0", "root", "netem", "delay",
		mean, variance, "distribution", opts.Distribution)
}

// Flaky introduces randomized packet loss.
func (IPTables) Flaky(ctx context.Context, node string) error {
	return ssh.Exec(ctx, node, "/sbin/tc", "qdisc", "add", "dev", "eth0", "root", "netem", "loss",
		"20%", "75%")
}

// Fast removes packet loss and delays.
func (IPTables) Fast(ctx context.Context, node string) error {
	output, err := ssh.CombinedOutput(ctx, node, "/sbin/tc", "qdisc", "del", "dev", "eth0", "root")
	if err != nil && strings.Contains(string(output), "RTNETLINK answers: No such file or directory") {
		err = nil
	}
	return err
}
