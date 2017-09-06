package core

import (
	"context"
	"fmt"
)

// DB allows Chaos to set up and tear down database.
type DB interface {
	// SetUp initializes the database.
	SetUp(ctx context.Context, node string) error
	// TearDown tears down the datase.
	TearDown(ctx context.Context, node string) error
	// Name returns the unique name for the database
	Name() string
}

// NoopDB is a DB but does nothing
type NoopDB struct {
}

// SetUp initializes the database.
func (NoopDB) SetUp(ctx context.Context, node string) error {
	return nil
}

// TearDown tears down the datase.
func (NoopDB) TearDown(ctx context.Context, node string) error {
	return nil
}

// Name returns the unique name for the database
func (NoopDB) Name() string {
	return "noop"
}

var dbs = map[string]DB{}

// RegisterDB registers db. Not thread-safe
func RegisterDB(db DB) {
	name := db.Name()
	_, ok := dbs[name]
	if ok {
		panic(fmt.Sprintf("%s is already registered", name))
	}

	dbs[name] = db
}

// GetDB gets the registered db. Panic if not found.
func GetDB(name string) DB {
	return dbs[name]
}

func init() {
	RegisterDB(NoopDB{})
}
