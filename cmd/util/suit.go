package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pingcap/chaos/pkg/control"
	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/nemesis"
	"github.com/pingcap/chaos/pkg/verify"
)

// Suit is a basic chaos testing suit with configurations to run chaos.
type Suit struct {
	*control.Config
	core.ClientCreator
	// nemesis, seperated by comma.
	Nemesises string

	VerifySuit verify.Suit
}

// Run runs the suit.
func (suit *Suit) Run(ctx context.Context, nodes []string) {
	var nemesisGens []core.NemesisGenerator
	for _, name := range strings.Split(suit.Nemesises, ",") {
		var g core.NemesisGenerator
		name := strings.TrimSpace(name)
		if len(name) == 0 {
			continue
		}

		switch name {
		case "random_kill", "all_kill", "minor_kill", "major_kill":
			g = nemesis.NewKillGenerator(suit.Config.DB, name)
		case "random_drop", "all_drop", "minor_drop", "major_drop":
			g = nemesis.NewDropGenerator(name)
		default:
			log.Fatalf("invalid nemesis generator %s", name)
		}

		nemesisGens = append(nemesisGens, g)
	}

	sctx, cancel := context.WithCancel(ctx)

	if len(nodes) != 0 {
		suit.Config.Nodes = nodes
	} else {
		// By default, we run TiKV/TiDB cluster on 5 nodes.
		for i := 1; i <= 5; i++ {
			name := fmt.Sprintf("n%d", i)
			suit.Config.Nodes = append(suit.Config.Nodes, name)
		}
	}

	c := control.NewController(
		sctx,
		suit.Config,
		suit.ClientCreator,
		nemesisGens,
		suit.VerifySuit,
	)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		c.Close()
		cancel()
	}()

	c.Run()
}
