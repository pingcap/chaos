package control

import (
	"testing"
	"time"

	"github.com/siddontang/chaos/pkg/core"
)

func TestControl(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	cfg := &Config{
		NodePort:     8080,
		RequestCount: 10,
		RunTime:      10 * time.Second,
		DB:           "noop",
		Client:       "noop",
	}

	c := NewController(cfg, core.NoopClientCreator{}, []core.NemesisGenerator{
		core.NoopNemesisGenerator{},
	})
	c.Run()
	c.Close()
}
