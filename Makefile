default: build

all: build 

build: node

node:
	go build -o bin/chaso-node cmd/node/main.go