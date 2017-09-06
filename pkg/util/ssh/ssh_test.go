package ssh

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestExec(t *testing.T) {
	ctx := context.Background()
	if err := Exec(ctx, "n1", "sleep 1"); err != nil {
		t.Fatalf("ssh failed %v", err)
	}

	if err := Exec(ctx, "n1", "cat non_exist_file"); err == nil {
		t.Fatal("ssh must fail")
	}

	if err := Exec(ctx, "n6", "sleep 1"); err == nil {
		t.Fatal("ssh must fail")
	}
}

func TestScp(t *testing.T) {
	dir, err := ioutil.TempDir(".", "var")
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(dir)

	f, err := os.Create(path.Join(dir, "a.log"))
	if err != nil {
		t.Fatalf("create file failed %v", err)
	}
	f.WriteString("Hello world")
	f.Close()

	ctx := context.Background()
	if err = Upload(ctx, path.Join(dir, "a.log"), "n1", "/tmp/b.log"); err != nil {
		t.Fatalf("upload file failed %v", err)
	}

	if err = Download(ctx, path.Join(dir, "b.log"), "n1", "/tmp/b.log"); err != nil {
		t.Fatalf("download file failed %v", err)
	}

	if err = Upload(ctx, dir, "n1", "/tmp/"); err != nil {
		t.Fatalf("upload folder failed %v", err)
	}

	if err = Download(ctx, path.Join(dir, "sub"), "n1", path.Join("/tmp", path.Base(dir))); err != nil {
		t.Fatalf("download foler failed %v", err)
	}
}
