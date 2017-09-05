package util

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestWget(t *testing.T) {
	var (
		name string
		err  error
	)
	for i := 0; i < 2; i++ {
		name, err = Wget(context.Background(), "https://raw.githubusercontent.com/pingcap/tikv/master/Cargo.toml", ".")
		if err != nil {
			t.Fatalf("download failed %v", err)
		}

		if !IsFileExist(name) {
			t.Fatalf("stat %s failed %v", name, err)
		}
	}

	os.Remove(name)
}

func TestInstallArchive(t *testing.T) {
	tmpDir, _ := ioutil.TempDir(".", "var")
	defer os.RemoveAll(tmpDir)

	err := InstallArchive(context.Background(), "https://github.com/siddontang/chaos/archive/master.zip", path.Join(tmpDir, "1"))
	if err != nil {
		t.Fatalf("install archive failed %v", err)
	}

	err = InstallArchive(context.Background(), "https://github.com/siddontang/chaos/archive/master.tar.gz", path.Join(tmpDir, "2"))
	if err != nil {
		t.Fatalf("install archive failed %v", err)
	}
}

func TestDaemon(t *testing.T) {
	t.Log("test may only be run in the chaos docker")

	tmpDir, _ := ioutil.TempDir(".", "var")
	defer os.RemoveAll(tmpDir)

	pidFile := path.Join(tmpDir, "sleep.pid")
	opts := NewDaemonOptions(tmpDir, pidFile, path.Join(tmpDir, "sleep.log"))
	err := StartDaemon(context.Background(), opts, "/bin/sleep", "100")
	if err != nil {
		t.Fatalf("start daemon failed %v", err)
	}

	pidStr := parsePID(pidFile)
	if pidStr == "" {
		t.Fatal("must have a pid file")
	}

	pid, _ := strconv.Atoi(pidStr)
	if !IsProcessExist(pid) {
		t.Fatalf("pid %d must exist", pid)
	}

	err = StopDaemon(context.Background(), "", pidFile)
	if err != nil {
		t.Fatalf("stop daemon failed %v", err)
	}

	if IsProcessExist(pid) {
		t.Fatalf("pid %d must not exist", pid)
	}

}
