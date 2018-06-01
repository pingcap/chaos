package tidb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/siddontang/chaos/pkg/core"
)

type multiBankClient struct {
	db         *sql.DB
	r          *rand.Rand
	accountNum int
}

func (c *multiBankClient) SetUp(ctx context.Context, nodes []string, node string) error {
	c.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s:4000)/test", node))
	if err != nil {
		return err
	}
	c.db = db

	db.SetMaxIdleConns(1 + c.accountNum)

	// Do SetUp in the first node
	if node != nodes[0] {
		return nil
	}

	log.Printf("begin to create table accounts on node %s", node)

	for i := 0; i < c.accountNum; i++ {
		sql := fmt.Sprintf(`create table if not exists accounts_%d
			(id     int not null primary key,
			balance bigint not null)`, i)

		if _, err = db.ExecContext(ctx, sql); err != nil {
			return err
		}

		sql = fmt.Sprintf("insert into accounts_%d values (?, ?)", i)
		if _, err = db.ExecContext(ctx, sql, i, initBalance); err != nil {
			return err
		}

	}

	return nil
}

func (c *multiBankClient) TearDown(ctx context.Context, nodes []string, node string) error {
	return c.db.Close()
}

func (c *multiBankClient) invokeRead(ctx context.Context, r bankRequest) bankResponse {
	txn, err := c.db.Begin()

	if err != nil {
		return bankResponse{Unknown: true}
	}
	defer txn.Rollback()

	var tso uint64
	if err = txn.QueryRow("select @@tidb_current_ts").Scan(&tso); err != nil {
		return bankResponse{Unknown: true}
	}

	balances := make([]int64, 0, c.accountNum)
	for i := 0; i < c.accountNum; i++ {
		var balance int64
		sql := fmt.Sprintf("select balance from accounts_%d", i)
		if err = txn.QueryRowContext(ctx, sql).Scan(&balance); err != nil {
			return bankResponse{Unknown: true}
		}
		balances = append(balances, balance)
	}

	return bankResponse{Balances: balances, Tso: tso}
}

func (c *multiBankClient) Invoke(ctx context.Context, node string, r interface{}) interface{} {
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
		tso         uint64
	)

	if err = txn.QueryRow("select @@tidb_current_ts").Scan(&tso); err != nil {
		return bankResponse{Ok: false}
	}

	if err = txn.QueryRowContext(ctx, fmt.Sprintf("select balance from accounts_%d where id = ? for update", arg.From), arg.From).Scan(&fromBalance); err != nil {
		return bankResponse{Ok: false}
	}

	if err = txn.QueryRowContext(ctx, fmt.Sprintf("select balance from accounts_%d where id = ? for update", arg.To), arg.To).Scan(&toBalance); err != nil {
		return bankResponse{Ok: false}
	}

	if fromBalance < arg.Amount {
		return bankResponse{Ok: false}
	}

	if _, err = txn.ExecContext(ctx, fmt.Sprintf("update accounts_%d set balance = balance - ? where id = ?", arg.From), arg.Amount, arg.From); err != nil {
		return bankResponse{Ok: false}
	}

	if _, err = txn.ExecContext(ctx, fmt.Sprintf("update accounts_%d set balance = balance + ? where id = ?", arg.To), arg.Amount, arg.To); err != nil {
		return bankResponse{Ok: false}
	}

	if err = txn.Commit(); err != nil {
		return bankResponse{Unknown: true, Tso: tso, FromBalance: fromBalance, ToBalance: toBalance}
	}

	return bankResponse{Ok: true, Tso: tso, FromBalance: fromBalance, ToBalance: toBalance}
}

func (c *multiBankClient) NextRequest() interface{} {
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

// MultiBankClientCreator creates a bank test client for tidb.
type MultiBankClientCreator struct {
}

// Create creates a client.
func (MultiBankClientCreator) Create(node string) core.Client {
	return &multiBankClient{
		accountNum: accountNum,
	}
}
