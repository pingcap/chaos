package util

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/siddontang/chaos/cmd/verifier/verify"
	"github.com/siddontang/chaos/pkg/control"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/nemesis"
)

// Suit is a basic chaos testing suit with configurations to run chaos.
type Suit struct {
	control.Config
	core.ClientCreator
	// model names, seperate by comma.
	ModelNames string
	// nemesis, seperated by comma.
	Nemesises string
}

// Run runs the suit.
func (suit *Suit) Run() {
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

	c := control.NewController(&suit.Config, suit.ClientCreator, nemesisGens)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigs
		c.Close()
		cancel()
	}()

	c.Run()

	verify.Verify(ctx, suit.Config.History, suit.ModelNames)
}
