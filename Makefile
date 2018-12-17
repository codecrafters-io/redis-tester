.PHONY: release build run run_with_redis

current_version = $(shell git describe --tags --abbrev=0)

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

bump_version:
	bumpversion --verbose --tag patch

upload_to_travis:
	aws s3 cp --acl public-read \
		s3://paul-redis-challenge/binaries/$(current_version)/$(current_version)_linux_amd64.tar.gz \
		s3://paul-redis-challenge/linux.tar.gz
	aws s3 cp --acl public-read \
		s3://paul-redis-challenge/binaries/$(current_version)/$(current_version)_darwin_amd64.tar.gz \
		s3://paul-redis-challenge/darwin.tar.gz
