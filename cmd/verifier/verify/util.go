package verify

import (
	"context"
	"log"
	"strings"

	"github.com/siddontang/chaos/db/tidb"
	"github.com/siddontang/chaos/pkg/check/porcupine"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
	"github.com/siddontang/chaos/pkg/model"
)

type suit struct {
	checker core.Checker
	model   core.Model
	parser  history.RecordParser
}

// Verify creates the verifier from model name and verfies the history file.
func Verify(ctx context.Context, historyFile string, modelNames string) {
	var suits []suit

	for _, name := range strings.Split(modelNames, ",") {
		s := suit{}
		switch name {
		case "tidb_bank":
			s.model, s.parser, s.checker = tidb.BankModel(), tidb.BankParser(), porcupine.Checker{}
		case "tidb_bank_tso":
			// Actually we can omit BankModel, since BankTsoChecker does not require a Model.
			s.model, s.parser, s.checker = tidb.BankModel(), tidb.BankParser(), tidb.BankTsoChecker()
		case "register":
			s.model, s.parser, s.checker = model.RegisterModel(), model.RegisterParser(), porcupine.Checker{}
		case "":
			continue
		default:
			log.Printf("%s is not supported", name)
			continue
		}

		suits = append(suits, s)
	}

	childCtx, cancel := context.WithCancel(ctx)

	go func() {
		for _, suit := range suits {
			log.Printf("begin to check %s with %s", suit.model.Name(), suit.checker.Name())
			ops, err := history.ReadHistory(historyFile, suit.parser)
			if err != nil {
				log.Fatalf("read history failed %v", err)
			}

			ops, err = history.CompleteOperations(ops, suit.parser)
			if err != nil {
				log.Fatalf("complete history failed %v", err)
			}

			ok, err := suit.checker.Check(suit.model, ops)
			if err != nil {
				log.Fatalf("verify history failed %v", err)
			}

			if !ok {
				log.Fatalf("%s: history %s is not linearizable", suit.model.Name(), historyFile)
			} else {
				log.Printf("%s: history %s is linearizable", suit.model.Name(), historyFile)
			}
		}

		cancel()
	}()

	<-childCtx.Done()
}
