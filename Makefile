.PHONY: release build run run_with_redis

release:
	goreleaser

build:
	go build -o dest/main.out

run: build
	dest/main.out --binary-path=./run_success.sh

run_debug: build
	dest/main.out --binary-path=./run_success.sh --debug=true

run_for_failure: build
	dest/main.out --binary-path=./run_failure.sh

run_with_redis: build
	dest/main.out --binary-path=redis-server
