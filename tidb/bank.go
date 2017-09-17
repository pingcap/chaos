package tidb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/anishathalye/porcupine"

	// use mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
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

func (c *bankClient) invokeRead(ctx context.Context, r bankRequest) bankResponse {
	rows, err := c.db.QueryContext(ctx, "select balance from accounts")
	if err != nil {
		return bankResponse{Unknown: true}
	}
	defer rows.Close()

	balances := make([]int64, 0, c.accountNum)
	for rows.Next() {
		var v int64
		if err = rows.Scan(&v); err != nil {
			return bankResponse{Unknown: true}
		}
		balances = append(balances, v)
	}

	return bankResponse{Balances: balances}
}

func (c *bankClient) Invoke(ctx context.Context, node string, r interface{}) interface{} {
	arg := r.(bankRequest)
	if arg.Op == 0 {
		return c.invokeRead(ctx, arg)
	}

	txn, err := c.db.Begin()

	if err != nil {
		return bankResponse{Ok: false}
	}
	defer txn.Rollback()

	var (
		fromBalance int64
		toBalance   int64
	)
	if err = txn.QueryRowContext(ctx, "select balance from accounts where id = ? for update", arg.From).Scan(&fromBalance); err != nil {
		return bankResponse{Ok: false}
	}

	if err = txn.QueryRowContext(ctx, "select balance from accounts where id = ? for update", arg.To).Scan(&toBalance); err != nil {
		return bankResponse{Ok: false}
	}

	if fromBalance < arg.Amount {
		return bankResponse{Ok: false}
	}

	if _, err = txn.ExecContext(ctx, "update accounts set balance = balance - ? where id = ?", arg.Amount, arg.From); err != nil {
		return bankResponse{Ok: false}
	}

	if _, err = txn.ExecContext(ctx, "update accounts set balance = balance + ? where id = ?", arg.Amount, arg.To); err != nil {
		return bankResponse{Ok: false}
	}

	if err = txn.Commit(); err != nil {
		return bankResponse{Unknown: true}
	}

	return bankResponse{Ok: true}
}

func (c *bankClient) NextRequest() interface{} {
	r := bankRequest{
		Op: c.r.Int() % 2,
	}
	if r.Op == 0 {
		return r
	}

	r.From = c.r.Intn(c.accountNum)

	r.To = c.r.Intn(c.accountNum)
	if r.From == r.To {
		r.To = (r.To + 1) % c.accountNum
	}

	r.Amount = 5
	return r
}

type bankRequest struct {
	// 0: read
	// 1: transfer
	Op     int
	From   int
	To     int
	Amount int64
}

type bankResponse struct {
	// read result
	Balances []int64
	// transfer ok or not
	Ok bool
	// read/transfer unknown
	Unknown bool
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

			if inp.Op == 0 {
				// read
				ok := out.Unknown || reflect.DeepEqual(st, out.Balances)
				return ok, state
			}

			// for transfer
			if !out.Ok && !out.Unknown {
				return true, state
			}

			newSt := append([]int64{}, st...)
			newSt[inp.From] -= inp.Amount
			newSt[inp.To] += inp.Amount
			return out.Ok || out.Unknown, newSt
		},

		Equal: func(state1, state2 interface{}) bool {
			return reflect.DeepEqual(state1, state2)
		},
	}
}

type bankParser struct {
}

func (p bankParser) OnRequest(data json.RawMessage) (interface{}, error) {
	r := bankRequest{}
	err := json.Unmarshal(data, &r)
	return r, err
}

func (p bankParser) OnResponse(data json.RawMessage) (interface{}, error) {
	r := bankResponse{}
	err := json.Unmarshal(data, &r)
	if r.Unknown {
		return nil, err
	}
	return r, err
}

func (p bankParser) OnNoopResponse() interface{} {
	return bankResponse{Unknown: true}
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

// BankVerifier verifies the bank history.
type BankVerifier struct {
}

// Verify verifies the bank history.
func (BankVerifier) Verify(name string) (bool, error) {
	return history.VerifyHistory(name, getBankModel(5), bankParser{})
}
