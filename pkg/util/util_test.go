package util

import (
	"context"
	"os"
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
	os.RemoveAll("./var")
	defer os.RemoveAll("./var")

	err := InstallArchive(context.Background(), "https://github.com/siddontang/chaos/archive/master.zip", "./var/1")
	if err != nil {
		t.Fatalf("install archive failed %v", err)
	}

	err = InstallArchive(context.Background(), "https://github.com/siddontang/chaos/archive/master.tar.gz", "./var/2")
	if err != nil {
		t.Fatalf("install archive failed %v", err)
	}

}
