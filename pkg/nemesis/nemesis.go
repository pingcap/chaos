package nemesis

import (
	"context"

	"github.com/siddontang/chaos/pkg/core"
)

type kill struct{}

func (kill) Invoke(ctx context.Context, node string, args ...string) error {
	db := core.GetDB(args[0])
	return db.Kill(ctx, node)
}

func (kill) Recover(ctx context.Context, node string, args ...string) error {
	db := core.GetDB(args[0])
	return db.Start(ctx, node)
}

func (kill) Name() string {
	return "kill"
}

func init() {
	core.RegisterNemesis(kill{})
}
