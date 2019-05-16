package tidb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pingcap/chaos/pkg/generator"
	"github.com/pingcap/chaos/pkg/history"

	"github.com/pingcap/chaos/pkg/core"
)

const (
	tablePrefix = "seq_"
	tpRead      = "read"
	tpWrite     = "write"
)

type request struct {
	Tp string
	K  int
}

type response struct {
	Ok      bool
	K       int
	V       []string
	Unknown bool
}

var (
	_ core.UnknownResponse = (*response)(nil)

	queue = struct {
		mu    sync.Mutex
		q     []int
		count int
		r     *rand.Rand
	}{
		mu: sync.Mutex{},
		q:  make([]int, 0, 10), // TODO: make the cap configurable.
		r:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
)

// IsUnknown implements UnknownResponse interface
func (r response) IsUnknown() bool {
	return r.Unknown
}

type sequentialClient struct {
	db         *sql.DB
	tableCount int
	keyCount   int
	gen        generator.Generator
}

// SequentialClientCreator creates a bank test client for tidb.
type SequentialClientCreator struct {
}

func genRequest() interface{} {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	c := cap(queue.q)
	l := len(queue.q)
	k := queue.count
	// The first len(queue.q) requests are all write requests.
	if k < c {
		goto WRITE
	} else if l == c {
		// Read if the queue is full
		goto READ
	} else if l <= 3 {
		// Write if the queue is empty
		goto WRITE
	} else if queue.r.Int()%2 == 1 {
		goto READ
	} else {
		goto WRITE
	}

READ:
	// Read
	// Pop the front
	k, queue.q = queue.q[0], append([]int{}, queue.q[1:]...)
	return request{Tp: tpRead, K: k}
WRITE:
	// Write
	// Push the end
	queue.count++
	queue.q = append(queue.q, k)
	return request{Tp: tpWrite, K: k}
}

// Create creates a new SequentialClient.
func (SequentialClientCreator) Create(node string) core.Client {
	return &sequentialClient{
		tableCount: 3,
		keyCount:   5,
		gen:        generator.Stagger(time.Millisecond*10, genRequest),
	}
}

func (c *sequentialClient) tableNames() []string {
	names := make([]string, 0, c.tableCount)
	for i := 0; i < c.tableCount; i++ {
		names = append(names, fmt.Sprintf("%s%d", tablePrefix, i))
	}
	return names
}

func subkeys(keyCount int, k int) []string {
	ks := make([]string, 0, keyCount)
	for i := 0; i < keyCount; i++ {
		ks = append(ks, fmt.Sprintf("%d_%d", k, i))
	}
	return ks
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}

func key2table(tableCount int, key string) string {
	return fmt.Sprintf("%s%d", tablePrefix, hash(key)%tableCount)
}

// SetUp sets up the client.
func (c *sequentialClient) SetUp(ctx context.Context, nodes []string, node string) error {
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

	tableNames := c.tableNames()
	log.Printf("begin to create table %v on node %s", tableNames, node)
	for _, tableName := range tableNames {
		log.Printf("try to drop table %s", tableName)
		if _, err = db.ExecContext(ctx,
			fmt.Sprintf(" drop table if exists %s", tableName)); err != nil {
			return err
		}
		sql := `create table if not exists %s (tkey varchar(255) primary key)`
		if _, err = db.ExecContext(ctx, fmt.Sprintf(sql, tableName)); err != nil {
			return err
		}
		log.Printf("created table %s", tableName)
	}

	return nil
}

// TearDown tears down the client.
func (c *sequentialClient) TearDown(ctx context.Context, nodes []string, node string) error {
	return c.db.Close()
}

// Invoke invokes a request to the database.
// Mostly, the return Response should implement UnknownResponse interface
func (c *sequentialClient) Invoke(ctx context.Context, node string, r interface{}) interface{} {
	req := r.(request)
	ks := subkeys(c.keyCount, req.K)
	if req.Tp == tpWrite {
		for _, k := range ks {
			sql := fmt.Sprintf("insert into %s values (?)", key2table(c.tableCount, k))
			_, err := c.db.ExecContext(ctx, sql, k)
			if err != nil {
				// TODO: retry on conflict
				log.Println(err)
				return response{Ok: false}
			}
		}
		return response{Ok: true}
	} else if req.Tp == tpRead {
		vs := make([]string, 0, len(ks))
		for i := len(ks) - 1; i >= 0; i-- {
			k := ks[i]
			sql := fmt.Sprintf("select tkey from %s where tkey = ?", key2table(c.tableCount, k))
			var v string
			err := c.db.QueryRowContext(ctx, sql, k).Scan(&v)
			if err != nil {
				// TODO: retry on conflict
				log.Println(err)
				return response{Ok: false}
			}
			vs = append(vs, v)
		}
		resp := response{
			Ok: true,
			K:  req.K,
			V:  vs,
		}
		return resp
	} else {
		panic(fmt.Sprintf("unknown req %v", req))
	}
}

// NextRequest generates a request for latter Invoke.
func (c *sequentialClient) NextRequest() interface{} {
	return c.gen()
}

// DumpState the database state(also the model's state)
func (c *sequentialClient) DumpState(ctx context.Context) (interface{}, error) {
	return nil, nil
}

// NewSequentialChecker returns a new sequentialChecker.
func NewSequentialChecker() core.Checker {
	return sequentialChecker{}
}

// sequentialChecker checks whether a history is sequential.
type sequentialChecker struct{}

// Check checks the sequential history.
func (sequentialChecker) Check(_ core.Model, ops []core.Operation) (bool, error) {
	for _, op := range ops {
		if op.Action == core.ReturnOperation {
			resp := op.Data.(response)
			if !resp.Unknown && resp.Ok {
				foundNoneEmpty := false
				for _, s := range resp.V {
					if !foundNoneEmpty {
						foundNoneEmpty = s != ""
					} else if s == "" {
						log.Printf("Find no sequential op %+v", resp)
						return false, nil
					}
				}
			}
		}
	}
	return true, nil
}

// Name returns the name of the verifier.
func (sequentialChecker) Name() string {
	return "sequential_checker"
}

type parser struct{}

// OnRequest impls history.RecordParser.OnRequest
func (p parser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := request{}
	err := json.Unmarshal(data, &r)
	return r, err
}

// OnResponse impls history.RecordParser.OnRequest
func (p parser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := response{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, err
}

// OnNoopResponse impls history.RecordParser.OnRequest
func (p parser) OnNoopResponse() interface{} {
	return response{Unknown: true}
}

func (p parser) OnState(data json.RawMessage) (interface{}, error) {
	return nil, nil
}

// NewSequentialParser returns a parser parses a history of bank operations.
func NewSequentialParser() history.RecordParser {
	return parser{}
}
