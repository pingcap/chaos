package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/siddontang/chaos/pkg/control"
	"github.com/siddontang/chaos/pkg/core"
	"github.com/siddontang/chaos/pkg/history"
	"github.com/siddontang/chaos/tidb"
)

var (
	nodePort     = flag.Int("node-port", 8080, "node port")
	requestCount = flag.Int("request-count", 1000, "client test request count")
	runTime      = flag.Duration("run-time", 10*time.Minute, "client test run time")
	clientCase   = flag.String("case", "bank", "client test case, like bank")
	historyFile  = flag.String("history", "./history.log", "history file")
)

func main() {
	flag.Parse()

	cfg := &control.Config{
		DB:           "tidb",
		NodePort:     *nodePort,
		RequestCount: *requestCount,
		RunTime:      *runTime,
		History:      *historyFile,
	}

	var (
		creator  core.ClientCreator
		verifier history.Verifier
	)

	switch *clientCase {
	case "bank":
		creator = tidb.BankClientCreator{}
		verifier = tidb.BankVerifier{}
	default:
		log.Fatalf("invalid client test case %s", *clientCase)
	}

	c := control.NewController(cfg, creator, nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		c.Close()
	}()

	c.Run()

	ok, err := verifier.Verify(*historyFile)
	if err != nil {
		log.Fatalf("verify history failed %v", err)
	}

	if !ok {
		log.Fatalf("%s history %s is not linearizable", *clientCase, *historyFile)
	}
}
