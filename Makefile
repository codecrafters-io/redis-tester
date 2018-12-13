.PHONY: release build run run_with_redis

release:
	goreleaser

build:
	go build -o dest/main.out

run: build
	dest/main.out --binary-path=./run.sh

run_with_redis: build
	dest/main.out --binary-path=redis-server
