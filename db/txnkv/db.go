package txnkv

import (
	"github.com/pingcap/chaos/db/cluster"
	"github.com/pingcap/chaos/pkg/core"
)

// db is the transactional KV database.
type db struct {
	cluster.Cluster
}

// Name returns the unique name for the database
func (db *db) Name() string {
	return "txnkv"
}

func init() {
	core.RegisterDB(&db{
		// TxnKV does not use TiDB.
		cluster.Cluster{IncludeTidb: false},
	})
}
