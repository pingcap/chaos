package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/siddontang/chaos/pkg/node"

	// register tidb
	_ "github.com/siddontang/chaos/tidb"
)

var (
	nodeAddr = flag.String("addr", ":8080", "node address")
	logFile  = flag.String("log-file", "/root/node.log", "node log file")
)

func main() {
	flag.Parse()

	f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("open log file failed %v", err)
		os.Exit(1)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Printf("begin to listen %s", *nodeAddr)
	n := node.NewNode(*nodeAddr)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		n.Run()
	}()

	<-sigs
	log.Printf("closing node")

	n.Close()

	log.Printf("node is closed")
}
