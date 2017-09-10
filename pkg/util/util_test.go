package util

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
	"time"
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

	archFile := path.Join(tmpDir, "a.tar.gz")
	testCreateArichive(t, path.Join(tmpDir, "test"), archFile)
	err = InstallArchive(context.Background(), "file://"+archFile, path.Join(tmpDir, "3"))
	if err != nil {
		t.Fatalf("install archive failed %v", err)
	}
}

func testCreateArichive(t *testing.T, srcDir string, name string) {
	os.MkdirAll(srcDir, 0755)
	f, err := os.Create(path.Join(srcDir, "a.log"))
	if err != nil {
		t.Fatalf("create file failed %v", err)
	}
	f.WriteString("hello world")
	f.Close()

	if err = exec.Command("tar", "-cf", name, "-C", srcDir, ".").Run(); err != nil {
		t.Fatalf("tar %s to %s failed %v", srcDir, name, err)
	}
}

func TestDaemon(t *testing.T) {
	t.Log("test may only be run in the chaos docker")

	tmpDir, _ := ioutil.TempDir(".", "var")
	defer os.RemoveAll(tmpDir)

	cmd := "/bin/sleep"
	pidFile := path.Join(tmpDir, "sleep.pid")
	opts := NewDaemonOptions(tmpDir, pidFile)
	err := StartDaemon(context.Background(), opts, cmd, "100")
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

	if !IsDaemonRunning(context.Background(), cmd, pidFile) {
		t.Fatal("daemon must be running")
	}

	err = StopDaemon(context.Background(), cmd, pidFile)
	if err != nil {
		t.Fatalf("stop daemon failed %v", err)
	}

	time.Sleep(time.Second)

	if IsProcessExist(pid) {
		t.Fatalf("pid %d must not exist", pid)
	}

	if IsFileExist(pidFile) {
		t.Fatalf("pid file must not exist")
	}

	if IsDaemonRunning(context.Background(), cmd, pidFile) {
		t.Fatal("daemon must be not running")
	}
}
