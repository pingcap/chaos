default: build

all: build 

build: node chaos

node:
	go build -o bin/chaos-node cmd/node/main.go

chaos:
	go build -o bin/chaos-tidb cmd/tidb/main.go