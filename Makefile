.PHONY: release build run run_with_redis

current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

docs:
	(sleep 0.5 && open http://localhost:6060/pkg/github.com/codecrafters-io/redis-tester/internal/)
	godoc -http=:6060

release:
	git tag v$(next_version_number)
	git push origin main v$(next_version_number)

build:
	go build -o dist/main.out ./cmd/tester

test:
	go test -p 1 -v ./internal/...

test_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"repl-master-cmd-prop\",\"tester_log_prefix\":\"replication-11\",\"title\":\"Replication Stage\"}]" \
	dist/main.out


test_tmp: build
	cd /tmp/abc && \
	CODECRAFTERS_SUBMISSION_DIR=/tmp/abc  \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"repl-master-cmd-prop\",\"tester_log_prefix\":\"replication-11\",\"title\":\"Replication Stage\"}]" \
	$(shell pwd)/dist/main.out

test_parallel: build
	rm -f dump.rdb
	CODECRAFTERS_SKIP_ANTI_CHEAT=true \
	CODECRAFTERS_TESTER_EXECUTABLE_PATH=$(shell pwd)/dist/main.out \
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"init\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage 1\"}, {\"slug\":\"ping-pong\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage 2\"}, {\"slug\":\"ping-pong-multiple\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage 3\"}]" \
	$(shell pwd)/dist/main.out
	rm -f dump.rdb

copy_course_file:
	gh api repos/codecrafters-io/build-your-own-redis/contents/course-definition.yml \
		| jq -r .content \
		| base64 -d \
		> internal/test_helpers/course_definition.yml

record_fixtures:
	CODECRAFTERS_RECORD_FIXTURES=true make test

update_tester_utils:
	go get -u github.com/codecrafters-io/tester-utils
