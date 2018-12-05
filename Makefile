default: build

all: build

build: chaos verifier

chaos: rawkv tidb txnkv

tidb:
	GO111MODULE=on go build -o bin/chaos-tidb cmd/tidb/main.go

rawkv:
	GO111MODULE=on go build -o bin/chaos-rawkv cmd/rawkv/main.go

txnkv:
	GO111MODULE=on go build -o bin/chaos-txnkv cmd/txnkv/main.go

verifier:
	GO111MODULE=on go build -o bin/chaos-verifier cmd/verifier/main.go