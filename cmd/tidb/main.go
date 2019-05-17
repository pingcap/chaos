package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/pingcap/chaos/cmd/util"
	"github.com/pingcap/chaos/db/tidb"
	"github.com/pingcap/chaos/pkg/check/porcupine"
	"github.com/pingcap/chaos/pkg/control"
	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/verify"
)

var (
	requestCount = flag.Int("request-count", 500, "client test request count")
	round        = flag.Int("round", 3, "client test request count")
	runTime      = flag.Duration("run-time", 10*time.Minute, "client test run time")
	clientCase   = flag.String("case", "bank", "client test case, like bank,multi_bank")
	historyFile  = flag.String("history", "./history.log", "history file")
	nemesises    = flag.String("nemesis", "", "nemesis, seperated by name, like random_kill,all_kill")
	checkerNames = flag.String("checker", "porcupine", "checker name, eg, porcupine, tidb_bank_tso")
)

func main() {
	flag.Parse()

	cfg := control.Config{
		DB:           "tidb",
		RequestCount: *requestCount,
		RunRound:     *round,
		RunTime:      *runTime,
		History:      *historyFile,
	}

	var creator core.ClientCreator
	var gen control.Generator
	switch *clientCase {
	case "bank":
		creator = tidb.BankClientCreator{}
		gen = tidb.BankGenRequest
	case "multi_bank":
		creator = tidb.MultiBankClientCreator{}
		gen = tidb.BankGenRequest
	default:
		log.Fatalf("invalid client test case %s", *clientCase)
	}

	var checker core.Checker
	switch *checkerNames {
	case "porcupine":
		checker = porcupine.Checker{}
	case "tidb_bank_tso":
		checker = tidb.BankTsoChecker()
	default:
		log.Fatalf("invalid checker %s", *checkerNames)
	}

	verifySuit := verify.Suit{
		Model:   tidb.BankModel(),
		Checker: checker,
		Parser:  tidb.BankParser(),
	}
	suit := util.Suit{
		Config:        &cfg,
		ClientCreator: creator,
		Generator:     gen,
		Nemesises:     *nemesises,
		VerifySuit:    verifySuit,
	}
	suit.Run(context.Background(), []string{})
}
