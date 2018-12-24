package tidb

import (
	"github.com/pingcap/chaos/db/cluster"
	"github.com/pingcap/chaos/pkg/core"
)

// db is the TiDB database.
type db struct {
	cluster.Cluster
}

// Name returns the unique name for the database
func (db *db) Name() string {
	return "tidb"
}

func init() {
	core.RegisterDB(&db{
		cluster.Cluster{IncludeTidb: true},
	})
}
