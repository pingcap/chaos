package core

import (
	"context"
	"fmt"
)

// Nemesis injects failure and disturbs the database.
type Nemesis interface {
	// // SetUp initializes the nemesis
	// SetUp(ctx context.Context, node string) error
	// // TearDown tears down the nemesis
	// TearDown(ctx context.Context, node string) error

	// Start starts the nemesis
	Start(ctx context.Context, node string, args ...string) error
	// Stop stops the nemesis
	Stop(ctx context.Context, node string, args ...string) error
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

// Start starts the nemesis
func (NoopNemesis) Start(ctx context.Context, node string, args ...string) error {
	return nil
}

// Stop stops the nemesis
func (NoopNemesis) Stop(ctx context.Context, node string, args ...string) error {
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
		panic(fmt.Sprintf("%s is already registered", name))
	}

	nemesises[name] = n
}

// GetNemesis gets the registered nemesis. Panic if not found.
func GetNemesis(name string) Nemesis {
	return nemesises[name]
}

func init() {
	RegisterNemesis(NoopNemesis{})
}
