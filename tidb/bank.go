package tidb

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

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

// TODO: now we just support nothing for request and response,
// but we will re-consider after we begin to implement linearizability check.
type bankRequest struct {
	from   int
	to     int
	amount int
}

type bankResponse struct {
}

func (r bankRequest) String() string {
	return fmt.Sprintf("%d %d %d", r.from, r.to, r.amount)
}

func (r bankResponse) String() string {
	return "ok"
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
