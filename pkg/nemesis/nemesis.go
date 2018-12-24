package nemesis

import (
	"context"

	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/util/net"
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

type drop struct {
	t net.IPTables
}

func (n drop) Invoke(ctx context.Context, node string, args ...string) error {
	for _, dropNode := range args {
		if node == dropNode {
			// Don't drop itself
			continue
		}

		if err := n.t.Drop(ctx, node, dropNode); err != nil {
			return err
		}
	}
	return nil
}

func (n drop) Recover(ctx context.Context, node string, args ...string) error {
	return n.t.Heal(ctx, node)
}

func (drop) Name() string {
	return "drop"
}

func init() {
	core.RegisterNemesis(kill{})
	core.RegisterNemesis(drop{})
}
