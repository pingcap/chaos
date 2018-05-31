package ssh

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
)

var (
	verbose = flag.Bool("ssh-verbose", false, "show the verbose of SSH command")
)

// Exec executes the cmd on the remote node.
// Here we assume we can run with `ssh node cmd` directly
// TODO: add a SSH config?
func Exec(ctx context.Context, node string, cmd string, args ...string) error {
	_, err := CombinedOutput(ctx, node, cmd, args...)
	return err
}

// CombinedOutput executes the cmd on the remote node and returns its combined standard
// output and standard error.
func CombinedOutput(ctx context.Context, node string, cmd string, args ...string) ([]byte, error) {
	v := []string{
		node,
		cmd,
	}
	v = append(v, args...)
	if *verbose {
		log.Printf("run %s %v on node %s", cmd, args, node)
	}
	data, err := exec.CommandContext(ctx, "ssh", v...).CombinedOutput()
	if err != nil {
		// For debug
		if *verbose {
			log.Printf("fail to run %v %q %v", v, data, err)
		}
	}
	return data, err
}

// Upload uploads files from local path to remote node path.
func Upload(ctx context.Context, localPath string, node string, remotePath string) error {
	return exec.CommandContext(ctx, "scp", "-r", localPath, fmt.Sprintf("%s:%s", node, remotePath)).Run()
}

// Download downloads files from remote node path to local path.
func Download(ctx context.Context, localPath string, node string, remotePath string) error {
	return exec.CommandContext(ctx, "scp", "-r", fmt.Sprintf("%s:%s", node, remotePath), localPath).Run()
}
