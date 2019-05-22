package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/pingcap/chaos/cmd/util"
	"github.com/pingcap/chaos/db/txnkv"
	"github.com/pingcap/chaos/pkg/check/porcupine"
	"github.com/pingcap/chaos/pkg/control"
	"github.com/pingcap/chaos/pkg/core"
	"github.com/pingcap/chaos/pkg/model"
	"github.com/pingcap/chaos/pkg/verify"
)

var (
	requestCount = flag.Int("request-count", 500, "client test request count")
	round        = flag.Int("round", 3, "client test request count")
	runTime      = flag.Duration("run-time", 10*time.Minute, "client test run time")
	clientCase   = flag.String("case", "register", "client test case, like register")
	historyFile  = flag.String("history", "./history.log", "history file")
	nemesises    = flag.String("nemesis", "", "nemesis, seperated by name, like random_kill,all_kill")
)

func main() {
	flag.Parse()

	cfg := control.Config{
		DB:           "txnkv",
		RequestCount: *requestCount,
		RunRound:     *round,
		RunTime:      *runTime,
		History:      *historyFile,
	}

	var creator core.ClientCreator
	var gen control.Generator
	switch *clientCase {
	case "register":
		creator = txnkv.RegisterClientCreator{}
		gen = txnkv.RegisterGenRequest
	default:
		log.Fatalf("invalid client test case %s", *clientCase)
	}

	verifySuit := verify.Suit{
		Model:   model.RegisterModel(),
		Checker: porcupine.Checker{},
		Parser:  model.RegisterParser(),
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
