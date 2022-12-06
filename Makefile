.PHONY: release build run run_with_redis

current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

release:
	git tag v$(next_version_number)
	git push origin master v$(next_version_number)

build:
	go build -o dist/main.out ./cmd/tester

test:
	go test -v ./internal/

test_starter_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	dist/starter.out

test_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_CURRENT_STAGE_SLUG="expiry" \
	dist/main.out

test_stage_1_failure: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/scenarios/bind/failure \
	CODECRAFTERS_CURRENT_STAGE_SLUG="init" \
	dist/main.out

test_ping_pong_eof: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/scenarios/ping-pong/eof \
	CODECRAFTERS_CURRENT_STAGE_SLUG="ping-pong" \
	dist/main.out

test_tmp: build
	cd /tmp/0d8e4ba11c57085f && \
	CODECRAFTERS_SUBMISSION_DIR=/tmp/0d8e4ba11c57085f  \
	CODECRAFTERS_CURRENT_STAGE_SLUG="ping-pong" \
	$(shell pwd)/dist/main.out

copy_course_file:
	hub api \
		repos/rohitpaulk/codecrafters-server/contents/codecrafters/store/data/redis.yml \
		| jq -r .content \
		| base64 -d \
		> internal/test_helpers/course_definition.yml

record_fixtures:
	CODECRAFTERS_RECORD_FIXTURES=true make test

update_tester_utils:
	go get -u github.com/codecrafters-io/tester-utils
