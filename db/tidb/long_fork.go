package tidb

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/history"
)

const (
	lfRead      = "read"
	lfWrite     = "write"
	lfGroupSize = 10
)

type lfRequest struct {
	Kind string
	// length=1 for write
	Keys []uint64
}
type lfResponse struct {
	Ok      bool
	Unknown bool
	Keys    []uint64
	Values  []uint64
}

var (
	lfState = struct {
		mu      sync.Mutex
		nextKey uint64
		workers map[string]uint64
	}{
		mu:      sync.Mutex{},
		nextKey: 0,
		workers: make(map[string]uint64),
	}
)

type longForkClient struct {
	db         *sql.DB
	r          *rand.Rand
	tableCount int
	node       string
}

func lfTableNames(tableCount int) []string {
	names := make([]string, 0, tableCount)
	for i := 0; i < tableCount; i++ {
		names = append(names, fmt.Sprintf("txn_lf_%d", i))
	}
	return names
}

func lfKey2Table(tableCount int, key uint64) string {
	b := make([]byte, 8)
	binary.PutUvarint(b, key)
	h := fnv.New32a()
	h.Write(b)
	hash := int(h.Sum32())
	return fmt.Sprintf("txn_lf_%d", hash%tableCount)
}

func (c *longForkClient) SetUp(ctx context.Context, nodes []string, node string) error {
	c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s:4000)/test", node))
	if err != nil {
		return err
	}
	c.db = db

	db.SetMaxIdleConns(1 + c.tableCount)

	// Do SetUp in the first node
	if node != nodes[0] {
		return nil
	}

	log.Printf("begin to create %v tables on node %s", c.tableCount, node)
	for _, tableName := range lfTableNames(c.tableCount) {
		log.Printf("try to drop table %s", tableName)
		if _, err = db.ExecContext(ctx,
			fmt.Sprintf("drop table if exists %s", tableName)); err != nil {
			return err
		}
		sql := "create table if not exists %s (id int not null primary key,sk int not null,val int)"
		if _, err = db.ExecContext(ctx, fmt.Sprintf(sql, tableName)); err != nil {
			return err
		}
		log.Printf("created table %s", tableName)
	}

	return nil
}

func (c *longForkClient) TearDown(ctx context.Context, nodes []string, node string) error {
	return c.db.Close()
}

func (c *longForkClient) Invoke(ctx context.Context, node string, r interface{}) interface{} {
	arg := r.(lfRequest)
	if arg.Kind == lfWrite {

		key := arg.Keys[0]
		sql := fmt.Sprintf("insert into %s (id, sk, val) values (?, ?, ?) on duplicate key update val = ?", lfKey2Table(c.tableCount, key))
		_, err := c.db.ExecContext(ctx, sql, key, key, 1, 1)
		if err != nil {
			return lfResponse{Ok: false, Keys: []uint64{key}, Values: []uint64{1}}
		}
		return lfResponse{Ok: true}

	} else if arg.Kind == lfRead {

		txn, err := c.db.Begin()
		defer txn.Rollback()
		if err != nil {
			return lfResponse{Ok: false}
		}

		values := make([]uint64, len(arg.Keys))
		for i, key := range arg.Keys {
			sql := fmt.Sprintf("select (val) from %s where id = ?", lfKey2Table(c.tableCount, key))
			err := txn.QueryRowContext(ctx, sql, key).Scan(&values[i])
			if err != nil {
				return lfResponse{Ok: false}
			}
		}

		if err = txn.Commit(); err != nil {
			return lfResponse{Unknown: true, Keys: arg.Keys, Values: values}
		}
		return lfResponse{Ok: true, Keys: arg.Keys, Values: values}

	} else {
		panic(fmt.Sprintf("unknown req %v", r))
	}
}

func (c *longForkClient) NextRequest() interface{} {
	lfState.mu.Lock()
	defer lfState.mu.Unlock()

	key, present := lfState.workers[c.node]
	if present {
		delete(lfState.workers, c.node)
		return lfRequest{Kind: lfRead, Keys: makeKeysInGroup(c.r, lfGroupSize, key)}
	}

	if c.r.Int()%2 == 0 {
		if size := len(lfState.workers); size > 0 {
			others := make([]uint64, size)
			idx := 0
			for _, value := range lfState.workers {
				others[idx] = value
				idx++
			}
			key := others[c.r.Intn(size)]
			return lfRequest{Kind: lfRead, Keys: makeKeysInGroup(c.r, lfGroupSize, key)}
		}
	}

	key = lfState.nextKey
	lfState.nextKey++
	lfState.workers[c.node] = key
	return lfRequest{Kind: lfWrite, Keys: []uint64{key}}
}

func (c *longForkClient) DumpState(ctx context.Context) (interface{}, error) {
	return nil, nil
}

func makeKeysInGroup(r *rand.Rand, groupSize uint64, key uint64) []uint64 {
	lower := key - key%groupSize
	base := r.Perm(int(groupSize))
	result := make([]uint64, groupSize)
	for i, num := range base {
		result[i] = uint64(num) + lower
	}
	return result
}

// LongForkClientCreator creates long fork test clients for tidb.
type LongForkClientCreator struct {
}

// Create creates a new longForkClient.
func (LongForkClientCreator) Create(node string) core.Client {
	return &longForkClient{
		tableCount: 7,
		node:       node,
	}
}

type lfParser struct{}

func (p lfParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := lfRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (p lfParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := lfResponse{}
	err := json.Unmarshal(data, &r)
	// I have no idea why we need this
	if r.Unknown {
		return nil, err
	}
	return r, err
}

func (p lfParser) OnNoopResponse() interface{} {
	return lfResponse{Unknown: true}
}

func (p lfParser) OnState(data json.RawMessage) (interface{}, error) {
	return nil, nil
}

// LongForkParser parses a history of long fork test.
func LongForkParser() history.RecordParser {
	return lfParser{}
}

type lfChecker struct{}

func ensureNoLongForks(ops []core.Operation) (bool, error) {
	groups := make(map[[lfGroupSize]uint64][][lfGroupSize]uint64)
	for _, op := range ops {
		if op.Action != core.ReturnOperation {
			continue
		}
		res := op.Data.(lfResponse)
		// you con not get request from the response...
		if len(res.Values) == 0 {
			// it's a write
			continue
		}
		if res.Ok && !res.Unknown {
			continue
		}
		if len(res.Keys) != lfGroupSize || len(res.Values) != lfGroupSize {
			return false, fmt.Errorf("The read respond should have %v keys and %v values, but it has %v keys and %v values",
				lfGroupSize, lfGroupSize, len(res.Keys), len(res.Values))
		}
		type pair struct {
			key   uint64
			value uint64
		}
		//sort key
		var pairs [lfGroupSize]pair
		for i := 0; i < lfGroupSize; i++ {
			pairs[i] = pair{key: res.Keys[i], value: res.Values[i]}
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].key < pairs[j].key })
		var keys [lfGroupSize]uint64
		var values [lfGroupSize]uint64
		for i := 0; i < lfGroupSize; i++ {
			keys[i] = pairs[i].key
			values[i] = pairs[i].value
		}
		groups[keys] = append(groups[keys], values)
	}
	for keys, results := range groups {
		count := len(results)
		for p := 0; p < count; p++ {
			for q := p + 1; q < count; q++ {
				values1 := results[p]
				values2 := results[q]
				//compare!
				var result int
				for i := 0; i < lfGroupSize; i++ {
					present1 := values1[i] > 0
					present2 := values2[i] > 0
					if present1 && !present2 {
						if result > 0 {
							log.Printf("Detected fork in history, read to %v returns %v and %v", keys, values1, values2)
							return false, nil
						}
						result = -1
					}
					if !present1 && present2 {
						if result < 0 {
							log.Printf("Detected fork in history, read to %v returns %v and %v", keys, values1, values2)
							return false, nil
						}
						result = 1
					}
					if present1 && present2 {
						if values1[i] != values2[i] {
							return false, fmt.Errorf("The key %v was write twice since it had two different values %v and %v",
								keys[i], values1[i], values2[i])
						}
					}
				}
			}
		}
	}
	return true, nil
}

func ensureNoMultipleWritesToOneKey(ops []core.Operation) (bool, error) {
	keySet := make(map[uint64]bool)
	for _, op := range ops {
		if op.Action != core.InvokeOperation {
			continue
		}
		req := op.Data.(lfRequest)
		if req.Kind != lfWrite {
			continue
		}
		for _, key := range req.Keys {
			if _, prs := keySet[key]; prs {
				return false, fmt.Errorf("The key %v was written twice", key)
			}
			keySet[key] = true
		}
	}
	return true, nil
}

func (lfChecker) Check(_ core.Model, ops []core.Operation) (bool, error) {
	if result, err := ensureNoMultipleWritesToOneKey(ops); err != nil {
		return false, err
	} else if !result {
		return false, nil
	}
	if result, err := ensureNoLongForks(ops); err != nil {
		return false, err
	} else if !result {
		return false, nil
	}
	return true, nil
}

func (lfChecker) Name() string {
	return "tidb_long_fork_checker"
}

// LongForkChecker checks the long fork test history.
func LongForkChecker() core.Checker {
	return lfChecker{}
}
