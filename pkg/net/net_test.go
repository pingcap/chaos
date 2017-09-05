package net

import (
	"context"
	"os"
	"testing"
)

func TestIsReachable(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	if !IsReachable(context.Background(), "n1") {
		t.Fatal("n1 must be reachable")
	}

	if IsReachable(context.Background(), "n0") {
		t.Fatal("n0 must not be reachable")
	}
}

func TestHostIP(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	if HostIP("n1") == "" {
		t.Fatal("must get a host IP for n1")
	}

	if HostIP("n0") != "" {
		t.Fatal("must not get a host IP for n0")
	}
}

func TestWget(t *testing.T) {
	for i := 0; i < 2; i++ {
		name, err := Wget(context.Background(), "https://raw.githubusercontent.com/pingcap/tikv/master/Cargo.toml")
		if err != nil {
			t.Fatalf("download failed %v", err)
		}

		_, err = os.Stat(name)
		if err != nil {
			t.Fatalf("stat %s failed %v", name, err)
		}
	}

	os.Remove(name)
}
