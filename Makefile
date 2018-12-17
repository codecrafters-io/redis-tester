.PHONY: release build run run_with_redis

release:
	rm -rf dist
	goreleaser

build:
	go build -o dist/main.out

run: build
	dist/main.out --binary-path=./run_success.sh

run_debug: build
	dist/main.out --binary-path=./run_success.sh --debug=true

run_for_failure: build
	dist/main.out --binary-path=./run_failure.sh

run_with_redis: build
	dist/main.out --binary-path=redis-server
