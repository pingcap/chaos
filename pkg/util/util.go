package util

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pingcap/chaos/pkg/util/ssh"
)

// IsFileExist runs on node and returns true if the file exists.
func IsFileExist(ctx context.Context, node string, name string) bool {
	err := ssh.Exec(ctx, node, "stat", name)
	return err == nil
}

// IsProcessExist runs on node and returns true if the porcess still exists.
func IsProcessExist(ctx context.Context, node string, pid int) bool {
	err := ssh.Exec(ctx, node, "kill", fmt.Sprintf("-s 0 %d", pid))
	return err == nil
}

// Wget runs on node, downloads a string URL to the dest directory and returns the file path.
// SKips if the file already exists.
func Wget(ctx context.Context, node string, rawURL string, dest string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if len(dest) == 0 {
		dest = "."
	}

	fileName := path.Base(u.Path)
	filePath := path.Join(dest, fileName)

	Mkdir(ctx, node, dest)
	err = ssh.Exec(ctx, node, "wget", "--tries", "20", "--waitretry", "60",
		"--retry-connrefused", "--dns-timeout", "60", "--connect-timeout", "60",
		"--read-timeout", "60", "--directory-prefix", dest, rawURL)
	return filePath, err
}

// InstallArchive runs on node, downloads the URL and extracts the archive to the dest diretory.
// Supports zip, and tarball.
func InstallArchive(ctx context.Context, node string, rawURL string, dest string) error {
	err := ssh.Exec(ctx, node, "mkdir", "-p", "/tmp/chaos")
	if err != nil {
		return err
	}

	tmpDir := fmt.Sprintf("/tmp/chaos/archive_%d", time.Now().UnixNano())
	Mkdir(ctx, node, tmpDir)
	defer RemoveDir(ctx, node, tmpDir)

	var name string
	if strings.HasPrefix(rawURL, "file://") {
		name = rawURL[len("file://"):]
	} else {
		if name, err = Wget(ctx, node, rawURL, "/tmp/chaos"); err != nil {
			return err
		}
	}

	if strings.HasSuffix(name, ".zip") {
		err = ssh.Exec(ctx, node, "unzip", "-d", tmpDir, name)
	} else {
		err = ssh.Exec(ctx, node, "tar", "-xf", name, "-C", tmpDir)
	}

	if err != nil {
		return err
	}

	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	RemoveDir(ctx, node, dest)
	Mkdir(ctx, node, path.Dir(dest))

	var files []string
	if files, err = ReadDir(ctx, node, tmpDir); err != nil {
		return err
	} else if len(files) == 1 && IsDir(ctx, node, path.Join(tmpDir, files[0])) {
		return ssh.Exec(ctx, node, "mv", path.Join(tmpDir, files[0]), dest)
	}

	return ssh.Exec(ctx, node, "mv", tmpDir, dest)
}

// ReadDir runs on node and lists the files of dir.
func ReadDir(ctx context.Context, node string, dir string) ([]string, error) {
	output, err := ssh.CombinedOutput(ctx, node, "ls", dir)
	if err != nil {
		return nil, err
	}

	seps := strings.Split(string(output), "\n")
	v := make([]string, 0, len(seps))
	for _, sep := range seps {
		sep = strings.TrimSpace(sep)
		if len(sep) > 0 {
			v = append(v, sep)
		}
	}

	return v, nil
}

// IsDir runs on node and checks path is directory or not.
func IsDir(ctx context.Context, node string, path string) bool {
	err := ssh.Exec(ctx, node, "test", "-d", path)
	return err == nil
}

// Mkdir runs on node and makes a directory
func Mkdir(ctx context.Context, node string, dir string) error {
	return ssh.Exec(ctx, node, "mkdir", "-p", dir)
}

// RemoveDir runs on node and removes the diretory
func RemoveDir(ctx context.Context, node string, dir string) error {
	return ssh.Exec(ctx, node, "rm", "-rf", dir)
}

// WriteFile runs on node and writes data to file
func WriteFile(ctx context.Context, node string, file string, data string) error {
	return ssh.Exec(ctx, node, "echo", "-e", data, ">", file)
}

// DaemonOptions is the options to start a command in daemon mode.
type DaemonOptions struct {
	ChDir   string
	PidFile string
	NoClose bool
}

// NewDaemonOptions returns a default daemon options.
func NewDaemonOptions(chDir string, pidFile string) DaemonOptions {
	return DaemonOptions{
		ChDir:   chDir,
		PidFile: pidFile,
		NoClose: false,
	}
}

// StartDaemon runs on node and starts a daemon process with options
func StartDaemon(ctx context.Context, node string, opts DaemonOptions, cmd string, cmdArgs ...string) error {
	var args []string
	args = append(args, "--start")
	args = append(args, "--background")
	if opts.NoClose {
		args = append(args, "--no-close")
	}
	args = append(args, "--make-pidfile")

	processName := path.Base(cmd)
	args = append(args, "--name", processName)

	args = append(args, "--pidfile", opts.PidFile)
	args = append(args, "--chdir", opts.ChDir)
	args = append(args, "--oknodo", "--startas", cmd)
	args = append(args, "--")
	args = append(args, cmdArgs...)

	return ssh.Exec(ctx, node, "start-stop-daemon", args...)
}

func parsePID(ctx context.Context, node string, pidFile string) string {
	data, err := ssh.CombinedOutput(ctx, node, "cat", pidFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// StopDaemon runs on node and stops the daemon process.
func StopDaemon(ctx context.Context, node string, cmd string, pidFile string) error {
	return stopDaemon(ctx, node, cmd, pidFile, "TERM")
}

// KillDaemon runs on node and kills the daemon process.
func KillDaemon(ctx context.Context, node string, cmd string, pidFile string) error {
	return stopDaemon(ctx, node, cmd, pidFile, "KILL")
}

func stopDaemon(ctx context.Context, node string, cmd string, pidFile string, sig string) error {
	name := path.Base(cmd)

	return ssh.Exec(ctx, node, "start-stop-daemon", "--stop", "--remove-pidfile",
		"--pidfile", pidFile, "--oknodo", "--name", name, "--signal", sig)
}

// IsDaemonRunning runs on node and returns whether the daemon is still running or not.
func IsDaemonRunning(ctx context.Context, node string, cmd string, pidFile string) bool {
	name := path.Base(cmd)

	err := ssh.Exec(ctx, node, "start-stop-daemon", "--status", "--pidfile", pidFile, "--name", name)

	return err == nil
}
