.PHONY: release build run run_with_redis

current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

release:
	git tag v$(next_version_number)
	git push origin main v$(next_version_number)

build:
	go build -o dist/main.out ./cmd/tester

test:
	go test -v ./internal/

test_starter_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	dist/starter.out

test_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"init\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage #1: Bind to a port\"},{\"slug\":\"ping-pong\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage #2: Respond to PING\"},{\"slug\":\"ping-pong-multiple\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage #3: Respond to multiple PINGs\"},{\"slug\":\"concurrent-clients\",\"tester_log_prefix\":\"stage-4\",\"title\":\"Stage #4: Handle concurrent clients\"},{\"slug\":\"echo\",\"tester_log_prefix\":\"stage-5\",\"title\":\"Stage #5: Implement the ECHO command\"},{\"slug\":\"set_get\",\"tester_log_prefix\":\"stage-6\",\"title\":\"Stage #6: Implement the SET \u0026 GET commands\"},{\"slug\":\"expiry\",\"tester_log_prefix\":\"stage-7\",\"title\":\"Stage #7: Expiry\"}]" \
	dist/main.out

test_tmp: build
	cd /tmp/0d8e4ba11c57085f && \
	CODECRAFTERS_SUBMISSION_DIR=/tmp/0d8e4ba11c57085f  \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"init\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage #1: Bind to a port\"},{\"slug\":\"ping-pong\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage #2: Respond to PING\"},{\"slug\":\"ping-pong-multiple\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage #3: Respond to multiple PINGs\"},{\"slug\":\"concurrent-clients\",\"tester_log_prefix\":\"stage-4\",\"title\":\"Stage #4: Handle concurrent clients\"},{\"slug\":\"echo\",\"tester_log_prefix\":\"stage-5\",\"title\":\"Stage #5: Implement the ECHO command\"},{\"slug\":\"set_get\",\"tester_log_prefix\":\"stage-6\",\"title\":\"Stage #6: Implement the SET \u0026 GET commands\"},{\"slug\":\"expiry\",\"tester_log_prefix\":\"stage-7\",\"title\":\"Stage #7: Expiry\"}]" \
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
