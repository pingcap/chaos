default: build

all: build 

build: chaos verifier

chaos:
	go build -o bin/chaos-tidb cmd/tidb/main.go

verifier:
	go build -o bin/chaos-verifier cmd/verifier/main.go