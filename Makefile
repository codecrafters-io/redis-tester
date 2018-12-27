.PHONY: release build run run_with_redis

current_version = $(shell git describe --tags --abbrev=0)

release:
	rm -rf dist
	goreleaser

build:
	go build -o dist/main.out

test: build
	dist/main.out --stage 8 --binary-path=./run_success.sh

test_first_stage: build
	dist/main.out --binary-path=./run_success.sh

test_debug: build
	dist/main.out --stage 8 --binary-path=./run_success.sh --debug=true

test_for_failure: build
	dist/main.out --stage 8 --binary-path=./run_failure.sh

test_with_redis: build
	dist/main.out --stage 8 --binary-path=redis-server

report_first_stage: build
	dist/main.out --binary-path=./run_success.sh --report --api-key=abcd

report: build
	dist/main.out --stage 8 --binary-path=./run_success.sh --report --api-key=abcd

report_with_redis: build
	dist/main.out --stage 8 --binary-path=redis-server --report --api-key=abcd

bump_version:
	bumpversion --verbose --tag patch

upload_to_travis:
	aws s3 cp --acl public-read \
		s3://paul-redis-challenge/binaries/$(current_version)/$(current_version)_linux_amd64.tar.gz \
		s3://paul-redis-challenge/linux.tar.gz
	aws s3 cp --acl public-read \
		s3://paul-redis-challenge/binaries/$(current_version)/$(current_version)_darwin_amd64.tar.gz \
		s3://paul-redis-challenge/darwin.tar.gz

bump_release_upload: bump_version release upload_to_travis
