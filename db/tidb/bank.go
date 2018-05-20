package tidb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/anishathalye/porcupine"

	// use mysql
	_ "github.com/go-sql-driver/mysql"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
)

const (
	accountNum  = 5
	initBalance = int64(1000)
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

	db.SetMaxIdleConns(1 + c.accountNum)

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
		if _, err = db.ExecContext(ctx, "insert into accounts values (?, ?)", i, initBalance); err != nil {
			return err
		}
	}

	return nil
}

func (c *bankClient) TearDown(ctx context.Context, nodes []string, node string) error {
	return c.db.Close()
}

func (c *bankClient) invokeRead(ctx context.Context, r bankRequest) bankResponse {
	txn, err := c.db.Begin()

	if err != nil {
		return bankResponse{Unknown: true}
	}
	defer txn.Rollback()

	var tso uint64
	if err = txn.QueryRow("select @@tidb_current_ts").Scan(&tso); err != nil {
		return bankResponse{Unknown: true}
	}

	rows, err := txn.QueryContext(ctx, "select balance from accounts")
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

	return bankResponse{Balances: balances, Tso: tso}
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
		tso         uint64
	)

	if err = txn.QueryRow("select @@tidb_current_ts").Scan(&tso); err != nil {
		return bankResponse{Ok: false}
	}

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
		return bankResponse{Unknown: true, Tso: tso, FromBalance: fromBalance, ToBalance: toBalance}
	}

	return bankResponse{Ok: true, Tso: tso, FromBalance: fromBalance, ToBalance: toBalance}
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
	// Transaction start timestamp
	Tso uint64
	// read result
	Balances []int64
	// transfer ok or not
	Ok bool
	// FromBalance is the previous from balance before transafer
	FromBalance int64
	// ToBalance is the previous to balance before transafer
	ToBalance int64
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
				v[i] = initBalance
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
		accountNum: accountNum,
	}
}

// BankVerifier verifies the bank history.
type BankVerifier struct {
}

// Verify verifies the bank history.
func (BankVerifier) Verify(historyFile string) (bool, error) {
	return history.VerifyHistory(historyFile, getBankModel(accountNum), bankParser{})
}

// BankTsoVerifier verifies the bank history.
// Unlike BankVerifier using porcupine, it uses a direct way because we know every timestamp of the transaction.
// So we can order all transactions with timetamp and replay them.
type BankTsoVerifier struct {
}

type tsoEvent struct {
	Tso uint64
	Op  int
	// For transfer
	From        int
	To          int
	FromBalance int64
	ToBalance   int64
	Amount      int64
	// For read
	Balances []int64

	Unknown bool
}

func (e tsoEvent) String() string {
	if e.Op == 0 {
		return fmt.Sprintf("%d, read %v, unknown %v", e.Tso, e.Balances, e.Unknown)
	}

	return fmt.Sprintf("%d, transafer %d %d(%d) -> %d(%d), unknown %v", e.Tso, e.Amount, e.From, e.FromBalance, e.To, e.ToBalance, e.Unknown)
}

type tsoEvents []*tsoEvent

func (s tsoEvents) Len() int           { return len(s) }
func (s tsoEvents) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s tsoEvents) Less(i, j int) bool { return s[i].Tso < s[j].Tso }

func parseTsoEvents(historyFile string) (tsoEvents, error) {
	events, err := history.ParseEvents(historyFile, bankParser{})
	if err != nil {
		return nil, err
	}

	return generateTsoEvents(events), nil
}

func generateTsoEvents(events []porcupine.Event) tsoEvents {
	tEvents := make(tsoEvents, 0, len(events))

	mapEvents := make(map[uint]porcupine.Event, len(events))
	for _, event := range events {
		if event.Kind == porcupine.CallEvent {
			mapEvents[event.Id] = event
			continue
		}

		// Handle Return Event
		// Find the corresponding Call Event
		callEvent, ok := mapEvents[event.Id]
		if !ok {
			continue
		}
		delete(mapEvents, event.Id)

		request := callEvent.Value.(bankRequest)
		response := event.Value.(bankResponse)

		if response.Tso == 0 {
			// We don't care operation which has no TSO.
			continue
		}

		tEvent := tsoEvent{
			Tso:     response.Tso,
			Op:      request.Op,
			Unknown: response.Unknown,
		}
		if request.Op == 0 {
			tEvent.Balances = response.Balances
		} else {
			tEvent.From = request.From
			tEvent.To = request.To
			tEvent.Amount = request.Amount
			tEvent.FromBalance = response.FromBalance
			tEvent.ToBalance = response.ToBalance
		}

		tEvents = append(tEvents, &tEvent)
	}
	sort.Sort(tEvents)
	return tEvents
}

// possibleBalances saves the possible balances we may meet.
// If the operation is Unknown, we can't know whether
// this operation is successful or not, so we may have different balance.
// E.g, the last balance is 1000, we reduce 100, so the next balance may be 1000 or 900 if Unknown.
type possibleBalances []int64

func (s possibleBalances) checkIn(balance int64) bool {
	for _, b := range s {
		if balance == b {
			return true
		}
	}

	return false
}

func verifyTsoEvents(events tsoEvents) bool {
	transferBalances := make([]possibleBalances, accountNum)
	readBalances := make([]possibleBalances, accountNum)

	// At first, the initialized balance is 1000.
	for i := 0; i < len(transferBalances); i++ {
		transferBalances[i] = []int64{initBalance}
		readBalances[i] = []int64{initBalance}
	}

	for _, event := range events {
		if event.Op == 1 {
			if !transferBalances[event.From].checkIn(event.FromBalance) {
				log.Printf("invalid event %s, last balances %v", event, transferBalances)
				return false
			}

			if !transferBalances[event.To].checkIn(event.ToBalance) {
				log.Printf("invalid event %s, last balances %v", event, transferBalances)
				return false
			}

			newFromBalance := event.FromBalance - event.Amount
			newToBalnce := event.ToBalance + event.Amount

			if !event.Unknown {
				transferBalances[event.From] = []int64{newFromBalance}
				transferBalances[event.To] = []int64{newToBalnce}
			} else {
				transferBalances[event.From] = []int64{event.FromBalance, newFromBalance}
				transferBalances[event.To] = []int64{event.ToBalance, newToBalnce}

			}

			// When we start a transfer at t1 (star timestamp), we can't know the exact commit timestamp (maybe t3),
			// so for every read transaction happends at [t1, t3], the transaction may read the old or the new value.
			readBalances[event.From] = []int64{event.FromBalance, newFromBalance}
			readBalances[event.To] = []int64{event.ToBalance, newToBalnce}
			// log.Printf("event: %s, new state %v", event, transferBalances)
		} else if event.Op == 0 {
			if event.Unknown {
				continue
			}

			sum := int64(0)
			for i, balance := range event.Balances {
				sum += balance
				if !readBalances[i].checkIn(balance) {
					log.Printf("invalid event %s, last balances %v", event, readBalances)
					return false
				}
			}

			if sum != int64(accountNum)*initBalance {
				log.Printf("invalid event %s, last balances %v", event, readBalances)
				return false
			}

			// log.Printf("event: %s, new state %v", event, readBalances)
		}
	}

	return true
}

// Verify verifes the bank history.
func (BankTsoVerifier) Verify(historyFile string) (bool, error) {
	events, err := parseTsoEvents(historyFile)
	if err != nil {
		return false, err
	}

	return verifyTsoEvents(events), nil
}
