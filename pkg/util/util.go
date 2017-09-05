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
// TODO: support local file
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
	if name, err = Wget(ctx, rawURL, tmpDir); err != nil {
		return err
	}

	if strings.HasSuffix(name, ".zip") {
		err = exec.CommandContext(ctx, "unzip", "-d", tmpDir, name).Run()
	} else {
		err = exec.CommandContext(ctx, "tar", "-xf", name, "-C", tmpDir).Run()
	}

	if err != nil {
		return err
	}

	// Remove the archive file
	os.Remove(name)

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
	Background       bool
	ChDir            string
	LogFile          string
	MakePidFile      bool
	MatchExecutable  bool
	MatchProcessName bool
	PidFile          string
	ProcessName      string
}

// NewDaemonOptions returns a default daemon options.
func NewDaemonOptions(chDir string, pidFile string, logFile string) DaemonOptions {
	return DaemonOptions{
		Background:       true,
		MakePidFile:      true,
		MatchExecutable:  true,
		MatchProcessName: false,
		ChDir:            chDir,
		PidFile:          pidFile,
		LogFile:          logFile,
	}
}

// StartDaemon starts a daemon process with options
func StartDaemon(ctx context.Context, opts DaemonOptions, cmd string, cmdArgs ...string) error {
	var args []string
	args = append(args, "--start")

	if opts.Background {
		args = append(args, "--background", "--no-close")
	}

	if opts.MakePidFile {
		args = append(args, "--make-pidfile")
	}

	if opts.MatchExecutable {
		args = append(args, "--exec", cmd)
	}

	if opts.MatchProcessName {
		processName := opts.ProcessName
		if len(processName) == 0 {
			processName = path.Base(cmd)
		}
		args = append(args, "--name", processName)
	}

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

// StopDaemon kills the daemon process by pid file or by the command name.
func StopDaemon(ctx context.Context, cmd string, pidFile string) error {
	var err error
	if len(cmd) > 0 {
		err = exec.CommandContext(ctx, "killall", "-9", "-w", cmd).Run()
	} else {
		if pid := parsePID(pidFile); len(pid) > 0 {
			err = exec.CommandContext(ctx, "kill", "-9", pid).Run()
		}
	}

	if err != nil {
		return err
	}

	if err = os.Remove(pidFile); os.IsNotExist(err) {
		err = nil
	}

	return nil
}
