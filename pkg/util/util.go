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
)

// IsFileExist returns true if the file exists.
func IsFileExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
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
