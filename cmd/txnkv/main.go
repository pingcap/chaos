package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/siddontang/chaos/cmd/util"
	"github.com/siddontang/chaos/db/txnkv"
	"github.com/siddontang/chaos/pkg/check/porcupine"
	"github.com/siddontang/chaos/pkg/control"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/model"
	"github.com/siddontang/chaos/pkg/verify"
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
	switch *clientCase {
	case "register":
		creator = txnkv.RegisterClientCreator{}
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
		Nemesises:     *nemesises,
		VerifySuit:    verifySuit,
	}
	suit.Run(context.Background(), []string{})
}
