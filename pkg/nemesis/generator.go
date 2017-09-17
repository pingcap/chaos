package nemesis

import (
	"math/rand"
	"time"

	"github.com/siddontang/chaos/pkg/core"
)

type randomKillGenerator struct {
	db string
}

func (g randomKillGenerator) Generate(nodes []string) []*core.NemesisOperation {
	index := rand.Intn(len(nodes))

	ops := make([]*core.NemesisOperation, len(nodes))

	ops[index] = &core.NemesisOperation{
		Name:    "kill",
		Args:    []string{g.db},
		RunTime: time.Second * time.Duration(rand.Intn(30)+1),
	}

	return ops
}

func (g randomKillGenerator) Name() string {
	return "random_kill"
}

// NewRandomKillGenerator kills db in one node randomly.
func NewRandomKillGenerator(db string) core.NemesisGenerator {
	return randomKillGenerator{db: db}
}

type allKillGenerator struct {
	db string
}

func (g allKillGenerator) Generate(nodes []string) []*core.NemesisOperation {
	ops := make([]*core.NemesisOperation, len(nodes))

	for i := 0; i < len(ops); i++ {
		ops[i] = &core.NemesisOperation{
			Name:    "kill",
			Args:    []string{g.db},
			RunTime: time.Second * time.Duration(rand.Intn(30)+1),
		}
	}
	return ops
}

func (g allKillGenerator) Name() string {
	return "all_kill"
}

// NewAllKillGenerator kills db in all nodes.
func NewAllKillGenerator(db string) core.NemesisGenerator {
	return allKillGenerator{db: db}
}
