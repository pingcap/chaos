package tidb

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/anishathalye/porcupine"

	// use mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/siddontang/chaos/pkg/core"
)

type bankClient struct {
	db         *sql.DB
	r          *rand.Rand
	accountNum int
}

func (c *bankClient) SetUp(ctx context.Context, nodes []string, node string) error {
	c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s:4000)/test", node))
	if err != nil {
		return err
	}
	c.db = db

	db.SetMaxIdleConns(2)

	// Do SetUp in the first node
	if node != nodes[0] {
		return nil
	}

	sql := `create table if not exists accounts
			(id     int not null primary key,
			balance bigint not null)`

	if _, err = db.ExecContext(ctx, sql); err != nil {
		return err
	}

	for i := 0; i < c.accountNum; i++ {
		if _, err = db.ExecContext(ctx, "insert into accounts values (?, 1000)", i); err != nil {
			return err
		}
	}

	return nil
}

func (c *bankClient) TearDown(ctx context.Context, nodes []string, node string) error {
	return c.db.Close()
}

func (c *bankClient) Invoke(ctx context.Context, node string, r core.Request) (core.Response, error) {
	arg := r.(bankRequest)
	txn, err := c.db.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	var (
		fromBalance int
		toBalance   int
	)
	if err = txn.QueryRowContext(ctx, "select balance from accounts where id = ? for update", arg.from).Scan(&fromBalance); err != nil {
		return nil, err
	}

	if err = txn.QueryRowContext(ctx, "select balance from accounts where id = ? for update", arg.to).Scan(&toBalance); err != nil {
		return nil, err
	}

	if fromBalance < arg.amount {
		return bankResponse{}, nil
	}

	if _, err = txn.ExecContext(ctx, "update accounts set balance = balance - ? where id = ?", arg.amount, arg.from); err != nil {
		return nil, err
	}

	if _, err = txn.ExecContext(ctx, "update accounts set balance = balance + ? where id = ?", arg.amount, arg.to); err != nil {
		return nil, err
	}

	if err = txn.Commit(); err != nil {
		return nil, err
	}

	return bankResponse{}, nil
}

func (c *bankClient) NextRequest() core.Request {
	var r bankRequest
	r.from = c.r.Intn(c.accountNum)

	r.to = c.r.Intn(c.accountNum)
	if r.from == r.to {
		r.to = (r.to + 1) % c.accountNum
	}

	r.amount = 5
	return r
}

type bankRequest struct {
	// 0: read
	// 1: transfer
	op     int
	from   int
	to     int
	amount int
}

type bankResponse struct {
	// read result
	balances []int64
	// transfer ok or not
	ok bool
	// read/transfer unknown
	unknown bool
}

func newBankEvent(v interface{}, id uint) porcupine.Event {
	if _, ok := v.(bankRequest); ok {
		return porcupine.Event{Kind: porcupine.CallEvent, Value: v, Id: id}
	}

	return porcupine.Event{Kind: porcupine.ReturnEvent, Value: v, Id: id}
}

func getBankModel(n int) porcupine.Model {
	return porcupine.Model{
		Init: func() interface{} {
			v := make([]int64, n)
			for i := 0; i < n; i++ {
				v[i] = 1000
			}
			return v
		},
		Step: func(state interface{}, input interface{}, output interface{}) (bool, interface{}) {
			st := state.([]int64)
			inp := input.(bankRequest)
			out := output.(bankResponse)

			if inp.op == 0 {
				// read
				ok := out.unknown || reflect.DeepEqual(st, out.balances)
				return ok, state
			}

			// for transfer
			newSt := append([]int64{}, st...)
			newSt[inp.from] -= int64(inp.amount)
			newSt[inp.to] += int64(inp.amount)
			return out.ok || out.unknown, newSt
		},
		
		Equal: func(state1, state2 interface{}) bool {
			return reflect.DeepEqual(state1, state2)
		},
	}
}

func (r bankRequest) String() string {
	if r.op == 0 {
		return "read"
	}

	return fmt.Sprintf("transfer %d %d %d", r.from, r.to, r.amount)
}

func (r bankResponse) String() string {
	return fmt.Sprintf("%v %v", r.balances, r.ok)
}

// BankClientCreator creates a bank test client for tidb.
type BankClientCreator struct {
}

// Create creates a client.
func (BankClientCreator) Create(node string) core.Client {
	return &bankClient{
		accountNum: 5,
	}
}
