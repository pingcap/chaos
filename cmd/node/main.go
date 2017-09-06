package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/siddontang/chaos/pkg/node"
)

var (
	nodeName = flag.String("name", "", "node name")
	nodeAddr = flag.String("addr", ":8080", "node address")
)

func main() {
	flag.Parse()

	n := node.NewNode(*nodeName, *nodeAddr)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		n.Run()
	}()

	<-sigs
	n.Close()
}
