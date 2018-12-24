package txnkv

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/anishathalye/porcupine"
	"github.com/pingcap/tidb/kv"
	"github.com/pingcap/tidb/store/tikv"
	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/model"
)

var (
	register = []byte("acc")

	closeOnce = sync.Once{}
)

type registerClient struct {
	db kv.Storage
	r  *rand.Rand
}

func (c *registerClient) SetUp(ctx context.Context, nodes []string, node string) error {
	c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	tikv.MaxConnectionCount = 128
	driver := tikv.Driver{}
	db, err := driver.Open(fmt.Sprintf("tikv://%s:2379?disableGC=true", node))
	if err != nil {
		return err
	}

	c.db = db

	// Do SetUp in the first node
	if node != nodes[0] {
		return nil
	}

	log.Printf("begin to initial register on node %s", node)

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err = tx.Set(register, []byte("0")); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (c *registerClient) TearDown(ctx context.Context, nodes []string, node string) error {
	var err error
	closeOnce.Do(func() {
		// It's a workaround for `panic: close of closed channel`.
		// `tikv.Driver.Open` will open the same instance if cluster id is
		// the same.
		//
		// See more: https://github.com/pingcap/tidb/blob/
		//           63c4562c27ad43165e6a0d5f890f33f3b1002b3f/store/tikv/kv.go#L95
		err = c.db.Close()
	})
	return err
}

func (c *registerClient) invokeRead(ctx context.Context, r model.RegisterRequest) model.RegisterResponse {
	tx, err := c.db.Begin()
	if err != nil {
		return model.RegisterResponse{Unknown: true}
	}
	defer tx.Rollback()

	val, err := tx.Get(register)
	if err != nil {
		return model.RegisterResponse{Unknown: true}
	}

	if err = tx.Commit(ctx); err != nil {
		return model.RegisterResponse{Unknown: true}
	}

	v, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return model.RegisterResponse{Unknown: true}
	}
	return model.RegisterResponse{Value: int(v)}
}

func (c *registerClient) invokeWrite(ctx context.Context, r model.RegisterRequest) model.RegisterResponse {
	tx, err := c.db.Begin()
	if err != nil {
		return model.RegisterResponse{Unknown: true}
	}
	defer tx.Rollback()

	val := fmt.Sprintf("%d", r.Value)
	if err = tx.Set(register, []byte(val)); err != nil {
		return model.RegisterResponse{Unknown: true}
	}

	if err = tx.Commit(ctx); err != nil {
		return model.RegisterResponse{Unknown: true}
	}
	return model.RegisterResponse{}
}

func (c *registerClient) Invoke(ctx context.Context, node string, r interface{}) interface{} {
	arg := r.(model.RegisterRequest)
	if arg.Op == model.RegisterRead {
		return c.invokeRead(ctx, arg)
	}
	return c.invokeWrite(ctx, arg)
}

func (c *registerClient) NextRequest() interface{} {
	r := model.RegisterRequest{
		Op: c.r.Intn(2) == 1,
	}
	if r.Op == model.RegisterRead {
		return r
	}

	r.Value = int(c.r.Int63())
	return r
}

// DumpState the database state(also the model's state)
func (c *registerClient) DumpState(ctx context.Context) (interface{}, error) {
	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	val, err := tx.Get(register)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	v, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func newRegisterEvent(v interface{}, id uint) porcupine.Event {
	if _, ok := v.(model.RegisterRequest); ok {
		return porcupine.Event{Kind: porcupine.CallEvent, Value: v, Id: id}
	}

	return porcupine.Event{Kind: porcupine.ReturnEvent, Value: v, Id: id}
}

// RegisterClientCreator creates a register test client for txnkv.
type RegisterClientCreator struct {
}

// Create creates a client.
func (RegisterClientCreator) Create(node string) core.Client {
	return &registerClient{}
}
