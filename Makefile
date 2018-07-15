default: build

all: build

build: chaos verifier

chaos: rawkv tidb txnkv

tidb:
	go build -o bin/chaos-tidb cmd/tidb/main.go

rawkv:
	go build -o bin/chaos-rawkv cmd/rawkv/main.go

txnkv:
	go build -o bin/chaos-txnkv cmd/txnkv/main.go

verifier:
	go build -o bin/chaos-verifier cmd/verifier/main.go