default: build

all: build 

build: agent chaos

agent:
	go build -o bin/chaos-agent cmd/agent/main.go

chaos:
	go build -o bin/chaos-tidb cmd/tidb/main.go

update:
	which glide >/dev/null || curl https://glide.sh/get | sh
	which glide-vc || go get -v -u github.com/sgotti/glide-vc
	glide update --strip-vendor --skip-test
	@echo "removing test files"
	glide vc --only-code --no-tests