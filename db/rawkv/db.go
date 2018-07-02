package rawkv

import (
	"github.com/siddontang/chaos/db/cluster"
	"github.com/siddontang/chaos/pkg/core"
)

// db is the TiDB database.
type db struct {
	cluster.Cluster
}

// Name returns the unique name for the database
func (db *db) Name() string {
	return "rawkv"
}

func init() {
	core.RegisterDB(&db{
		// RawKV does not use TiDB.
		cluster.Cluster{IncludeTidb: false},
	})
}
