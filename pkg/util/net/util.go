package net

import (
	"context"
	"net"
	"os/exec"
)

// IsReachable returns the destNode is reachable with ping.
func IsReachable(ctx context.Context, destNode string) bool {
	cmd := exec.CommandContext(ctx, "ping", "-w", "1", destNode)
	return cmd.Run() == nil
}

// HostIP looks up an IP for a hostname.
func HostIP(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	return ips[0].String()
}
