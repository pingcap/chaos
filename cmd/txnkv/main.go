package main

import (
	"flag"
	"log"
	"time"

	"github.com/siddontang/chaos/cmd/util"
	"github.com/siddontang/chaos/db/txnkv"
	"github.com/siddontang/chaos/pkg/control"
	"github.com/siddontang/chaos/pkg/core"
)

var (
	requestCount = flag.Int("request-count", 500, "client test request count")
	runTime      = flag.Duration("run-time", 10*time.Minute, "client test run time")
	clientCase   = flag.String("case", "register", "client test case, like bank,multi_bank")
	historyFile  = flag.String("history", "./history.log", "history file")
	nemesises    = flag.String("nemesis", "", "nemesis, seperated by name, like random_kill,all_kill")
)

func main() {
	flag.Parse()

	cfg := control.Config{
		DB:           "txnkv",
		RequestCount: *requestCount,
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

	suit := util.Suit{
		Config:        cfg,
		ClientCreator: creator,
		ModelNames:    "register",
		Nemesises:     *nemesises,
	}
	suit.Run()
}
