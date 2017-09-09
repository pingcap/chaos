default: build

all: build 

build: node

node:
	go build -o bin/chaos-node cmd/node/main.go