package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/siddontang/chaos/cmd/verifier/verify"
)

var (
	historyFile = flag.String("history", "./history.log", "history file")
	names       = flag.String("names", "", "verifier names, seperate by comma")
	pprofAddr   = flag.String("pprof", "0.0.0.0:6060", "Pprof address")
)

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		http.ListenAndServe(*pprofAddr, nil)
	}()

	go func() {
		<-sigs
		cancel()
	}()

	verify.Verify(ctx, *historyFile, *names)
}
