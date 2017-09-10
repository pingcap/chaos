package main

import (
	"log"
	"strings"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/siddontang/chaos/pkg/node"
)

var (
	nodeName = flag.String("name", "", "name for the current node")
	nodeList = flag.String("nodes", "n1,n2,n3,n4,n5", "cluster node names")
	nodeAddr = flag.String("addr", ":8080", "node address")
)

func main() {
	flag.Parse()

	nodes := strings.Split(*nodeList, ",")
	found := false 
	for i := 0; i < len(nodes); i++ {
		if nodes[i] == *nodeName {
			found = true 
			break 
		}	
	}

	if !found {
		log.Fatalf("missing node name %s in the cluster node names %v", *nodeName, nodes)
	}

	n := node.NewNode(nodes, *nodeName, *nodeAddr)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		n.Run()
	}()

	<-sigs
	n.Close()
}
