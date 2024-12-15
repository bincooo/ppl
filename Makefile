TARGET_EXEC := server

.PHONY: all changelog clean install build

all: clean install build-linux build-linux-freebsd

changelog:
	conventional-changelog -p angular -o CHANGELOG.md -w -r 0

clean:
	go clean -cache

install: clean
	go install -ldflags="-s -w" -trimpath ./cmd/iocgo

build-linux:
	GOARCH=amd64 GOOS=linux go build -toolexec iocgo -ldflags="-s -w" -o bin/linux/${TARGET_EXEC} -trimpath main.go

build-linux-freebsd:
	GOOS=freebsd GOARCH=amd64 go build --toolexec "iocgo" -ldflags="-s -w" -o bin/linux/${TARGET_EXEC}-freebsd -trimpath main.go