package util

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// IsFileExist returns true if the file exists.
func IsFileExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// IsProcessExist returns true if the porcess still exists.
func IsProcessExist(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// Wget downloads a string URL to the dest directory and returns the file path.
// SKips if the file already exists.
func Wget(ctx context.Context, rawURL string, dest string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if len(dest) == 0 {
		dest = "."
	}

	fileName := path.Base(u.Path)
	filePath := path.Join(dest, fileName)
	if IsFileExist(filePath) {
		return filePath, nil
	}

	os.MkdirAll(dest, 0755)
	err = exec.CommandContext(ctx, "wget", "--tries", "20", "--waitretry", "60",
		"--retry-connrefused", "--dns-timeout", "60", "--connect-timeout", "60",
		"--read-timeout", "60", "--directory-prefix", dest, rawURL).Run()
	return filePath, err
}

// InstallArchive downloads the URL and extracts the archive to the dest diretory.
// Supports zip, and tarball.
func InstallArchive(ctx context.Context, rawURL string, dest string) error {
	err := os.MkdirAll("/tmp/chaos", 0755)
	if err != nil {
		return err
	}

	var tmpDir string
	if tmpDir, err = ioutil.TempDir("/tmp/chaos", "archive_"); err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	var name string
	if strings.HasPrefix(rawURL, "file://") {
		name = strings.Trim(rawURL, "file://")
	} else {
		if name, err = Wget(ctx, rawURL, "/tmp/chaos"); err != nil {
			return err
		}
	}

	if strings.HasSuffix(name, ".zip") {
		err = exec.CommandContext(ctx, "unzip", "-d", tmpDir, name).Run()
	} else {
		err = exec.CommandContext(ctx, "tar", "-xf", name, "-C", tmpDir).Run()
	}

	if err != nil {
		return err
	}

	if dest, err = filepath.Abs(dest); err != nil {
		return err
	}

	os.RemoveAll(dest)
	os.MkdirAll(path.Dir(dest), 0755)

	var files []os.FileInfo
	if files, err = ioutil.ReadDir(tmpDir); err != nil {
		return err
	} else if len(files) == 1 && files[0].IsDir() {
		return os.Rename(path.Join(tmpDir, files[0].Name()), dest)
	}

	return os.Rename(tmpDir, dest)
}

// DaemonOptions is the options to start a command in daemon mode.
type DaemonOptions struct {
	ChDir   string
	LogFile string
	PidFile string
}

// NewDaemonOptions returns a default daemon options.
func NewDaemonOptions(chDir string, pidFile string, logFile string) DaemonOptions {
	return DaemonOptions{
		ChDir:   chDir,
		PidFile: pidFile,
		LogFile: logFile,
	}
}

// StartDaemon starts a daemon process with options
func StartDaemon(ctx context.Context, opts DaemonOptions, cmd string, cmdArgs ...string) error {
	var args []string
	args = append(args, "--start")
	args = append(args, "--background", "--no-close")
	args = append(args, "--make-pidfile")

	processName := path.Base(cmd)
	args = append(args, "--name", processName)

	args = append(args, "--pidfile", opts.PidFile)
	args = append(args, "--chdir", opts.ChDir)
	args = append(args, "--oknodo", "--startas", cmd)
	args = append(args, "--")
	args = append(args, cmdArgs...)

	c := exec.CommandContext(ctx, "start-stop-daemon", args...)
	logFile, err := os.Create(opts.LogFile)
	if err != nil {
		return err
	}
	defer logFile.Close()
	c.Stdout = logFile
	c.Stderr = logFile

	return c.Run()
}

func parsePID(pidFile string) string {
	data, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// StopDaemon stops the daemon process.
func StopDaemon(ctx context.Context, cmd string, pidFile string) error {
	return stopDaemon(ctx, cmd, pidFile, "TERM")
}

// KillDaemon kills the daemon process.
func KillDaemon(ctx context.Context, cmd string, pidFile string) error {
	return stopDaemon(ctx, cmd, pidFile, "KILL")
}

func stopDaemon(ctx context.Context, cmd string, pidFile string, sig string) error {
	name := path.Base(cmd)

	return exec.CommandContext(ctx, "start-stop-daemon", "--stop", "--remove-pidfile",
		"--pidfile", pidFile, "--oknodo", "--name", name, "--signal", sig).Run()
}

// IsDaemonRunning returns whether the daemon is still running or not.
func IsDaemonRunning(ctx context.Context, cmd string, pidFile string) bool {
	name := path.Base(cmd)

	err := exec.CommandContext(ctx, "start-stop-daemon", "--status", "--pidfile", pidFile, "--name", name).Run()

	return err == nil
}
