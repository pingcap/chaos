package ssh

import (
	"context"
	"fmt"
	"os/exec"
)

// Exec executes the cmd on the remote node.
// Here we assume we can run with `ssh node cmd` directly
// TODO: add a SSH config?
func Exec(ctx context.Context, node string, cmd string) error {
	return exec.CommandContext(ctx, "ssh", node, cmd).Run()
}

// Upload uploads files from local path to remote node path.
func Upload(ctx context.Context, localPath string, node string, remotePath string) error {
	return exec.CommandContext(ctx, "scp", "-r", localPath, fmt.Sprintf("%s:%s", node, remotePath)).Run()
}

// Download downloads files from remote node path to local path.
func Download(ctx context.Context, localPath string, node string, remotePath string) error {
	return exec.CommandContext(ctx, "scp", "-r", fmt.Sprintf("%s:%s", node, remotePath), localPath).Run()
}
