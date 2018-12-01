package control

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
	"github.com/siddontang/chaos/pkg/verify"
)

func TestControl(t *testing.T) {
	t.Log("test can only be run in the chaos docker")

	cfg := &Config{
		RequestCount: 10,
		RunTime:      10 * time.Second,
		RunRound:     3,
		DB:           "noop",
		History:      "/tmp/chaos/a.log",
	}

	defer os.Remove("/tmp/chaos/a.log")

	ngs := []core.NemesisGenerator{
		core.NoopNemesisGenerator{},
	}
	client := core.NoopClientCreator{}

	verifySuit := verify.Suit{
		Model:   &core.NoopModel{},
		Checker: core.NoopChecker{},
		Parser:  history.NoopParser{},
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := NewController(ctx, cfg, client, ngs, verifySuit)
	c.Run()
	c.Close()
	cancel()
}
