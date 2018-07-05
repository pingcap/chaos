default: build

all: build

build: chaos verifier

chaos: rawkv tidb

tidb:
	go build -o bin/chaos-tidb cmd/tidb/main.go

rawkv:
	go build -o bin/chaos-rawkv cmd/rawkv/main.go

verifier:
	go build -o bin/chaos-verifier cmd/verifier/main.go