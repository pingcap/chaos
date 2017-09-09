package control

import (
	"testing"
	"time"

	"github.com/siddontang/chaos/pkg/core"
)

func TestControl(t *testing.T) {
	cfg := &Config{
		NodePort:     8080,
		RequestCount: 10,
		RunTime:      10 * time.Second,
		DB:           "noop",
	}

	c := NewController(cfg, core.NoopClientCreator{}, []core.NemesisGenerator{
		core.NoopNemesisGenerator{},
	})
	c.Run()
	c.Close()
}
