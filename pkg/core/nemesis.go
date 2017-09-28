package core

import (
	"context"
	"fmt"
	"time"
)

// Nemesis injects failure and disturbs the database.
// Nemesis is used in node, you can define your own nemesis and register it.
type Nemesis interface {
	// // SetUp initializes the nemesis
	// SetUp(ctx context.Context, node string) error
	// // TearDown tears down the nemesis
	// TearDown(ctx context.Context, node string) error

	// Invoke executes the nemesis
	Invoke(ctx context.Context, node string, args ...string) error
	// Recover recovers the nemesis
	Recover(ctx context.Context, node string, args ...string) error
	// Name returns the unique name for the nemesis
	Name() string
}

// NoopNemesis is a nemesis but does nothing
type NoopNemesis struct {
}

// // SetUp initializes the nemesis
// func (NoopNemesis) SetUp(ctx context.Context, node string) error {
// 	return nil
// }

// // TearDown tears down the nemesis
// func (NoopNemesis) TearDown(ctx context.Context, node string) error {
// 	return nil
// }

// Invoke executes the nemesis
func (NoopNemesis) Invoke(ctx context.Context, node string, args ...string) error {
	return nil
}

// Recover recovers the nemesis
func (NoopNemesis) Recover(ctx context.Context, node string, args ...string) error {
	return nil
}

// Name returns the unique name for the nemesis
func (NoopNemesis) Name() string {
	return "noop"
}

var nemesises = map[string]Nemesis{}

// RegisterNemesis registers nemesis. Not thread-safe.
func RegisterNemesis(n Nemesis) {
	name := n.Name()
	_, ok := nemesises[name]
	if ok {
		panic(fmt.Sprintf("nemesis %s is already registered", name))
	}

	nemesises[name] = n
}

// GetNemesis gets the registered nemesis.
func GetNemesis(name string) Nemesis {
	return nemesises[name]
}

// NemesisOperation is nemesis operation used in control.
type NemesisOperation struct {
	// Nemesis name
	Name string
	// Nemesis invoke args
	InvokeArgs []string
	// Nemesis recover args
	RecoverArgs []string
	// Nemesis execute time
	RunTime time.Duration
}

// NemesisGenerator is used in control, it will generate a nemesis operation
// and then the control can use it to disturb the cluster.
type NemesisGenerator interface {
	// Generate generates the nemesis operation for all nodes.
	// Every node will be assigned a nemesis operation.
	Generate(nodes []string) []*NemesisOperation
	Name() string
}

// NoopNemesisGenerator generates
type NoopNemesisGenerator struct {
}

// Name returns the name
func (NoopNemesisGenerator) Name() string {
	return "noop"
}

//Generate generates the nemesis operation for the nodes.
func (NoopNemesisGenerator) Generate(nodes []string) []*NemesisOperation {
	ops := make([]*NemesisOperation, len(nodes))
	for i := 0; i < len(ops); i++ {
		ops[i] = &NemesisOperation{
			Name:        "noop",
			InvokeArgs:  nil,
			RecoverArgs: nil,
			RunTime:     0,
		}
	}
	return ops
}

func init() {
	RegisterNemesis(NoopNemesis{})
}
