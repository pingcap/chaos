package control

import (
	"os"
	"testing"
	"time"

	"github.com/siddontang/chaos/pkg/core"
)

func TestControl(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	cfg := &Config{
		RequestCount: 10,
		RunTime:      10 * time.Second,
		DB:           "noop",
		History:      "/tmp/chaos/a.log",
	}

	defer os.Remove("/tmp/chaos/a.log")

	c := NewController(cfg, core.NoopClientCreator{}, []core.NemesisGenerator{
		core.NoopNemesisGenerator{},
	})
	c.Run()
	c.Close()
}
