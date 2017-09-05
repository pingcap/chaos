package net

import (
	"context"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
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

// Wget downloads a string URL and returns the filename as a string.
// SKips if the file already exists.
func Wget(ctx context.Context, rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	filleName := path.Base(u.Path)
	if _, err := os.Stat(filleName); err == nil {
		return filleName, nil
	}

	err = exec.CommandContext(ctx, "wget", "--tries", "20", "--waitretry", "60",
		"--retry-connrefused", "--dns-timeout", "60", "--connect-timeout", "60",
		"--read-timeout", "60", rawURL).Run()
	return filleName, err
}
