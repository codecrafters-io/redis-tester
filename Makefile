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
	go test -v ./internal/

test_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"init\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage #1: Bind to a port\"},{\"slug\":\"ping-pong\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage #2: Respond to PING\"},{\"slug\":\"ping-pong-multiple\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage #3: Respond to multiple PINGs\"},{\"slug\":\"concurrent-clients\",\"tester_log_prefix\":\"stage-4\",\"title\":\"Stage #4: Handle concurrent clients\"},{\"slug\":\"echo\",\"tester_log_prefix\":\"stage-5\",\"title\":\"Stage #5: Implement the ECHO command\"},{\"slug\":\"set_get\",\"tester_log_prefix\":\"stage-6\",\"title\":\"Stage #6: Implement the SET \u0026 GET commands\"},{\"slug\":\"expiry\",\"tester_log_prefix\":\"stage-7\",\"title\":\"Stage #7: Expiry\"},{\"slug\":\"repl-custom-port\",\"tester_log_prefix\":\"stage-101\",\"title\":\"Stage #101: Replication - Custom Port\"},{\"slug\":\"repl-custom-port\",\"tester_log_prefix\":\"stage-101\",\"title\":\"Stage #101: Replication - Custom Port\"}, {\"slug\":\"repl-info\",\"tester_log_prefix\":\"stage-102\",\"title\":\"Stage #102: Replication - Info on Master\"}, {\"slug\":\"repl-info-replica\",\"tester_log_prefix\":\"stage-103\",\"title\":\"Stage #103: Replication - Info on Replica\"}, {\"slug\":\"repl-id\",\"tester_log_prefix\":\"stage-104\",\"title\":\"Stage #104: Replication - Replication ID and Offset\"}, {\"slug\":\"repl-replica-ping\",\"tester_log_prefix\":\"stage-105\",\"title\":\"Stage #105: Replication - Handshake 1\"},{\"slug\":\"repl-replica-replconf\",\"tester_log_prefix\":\"stage-106\",\"title\":\"Stage #106: Replication - Handshake 2\"},{\"slug\":\"repl-replica-psync\",\"tester_log_prefix\":\"stage-107\",\"title\":\"Stage #107: Replication - Handshake 3\"},{\"slug\":\"repl-master-replconf\",\"tester_log_prefix\":\"stage-108\",\"title\":\"Stage #108: Replication - REPLCONF\"},{\"slug\":\"repl-master-psync\",\"tester_log_prefix\":\"stage-109\",\"title\":\"Stage #109: Replication - PSYNC\"},{\"slug\":\"repl-master-psync-rdb\",\"tester_log_prefix\":\"stage-110\",\"title\":\"Stage #110: Replication - PSYNC w RDB file\"},{\"slug\":\"repl-master-cmd-prop\",\"tester_log_prefix\":\"stage-111\",\"title\":\"Stage #111: Command Propagation\"},{\"slug\":\"repl-cmd-processing\",\"tester_log_prefix\":\"stage-112\",\"title\":\"Stage #112: Command Processing\"},{\"slug\":\"repl-final-boss\",\"tester_log_prefix\":\"stage-113\",\"title\":\"Stage #113: Support multiple Replicas\"}]" \
	dist/main.out


test_tmp: build
	cd /workspaces/redis-go/codecrafters-redis-go/ && \
	CODECRAFTERS_SUBMISSION_DIR=/workspaces/redis-go/codecrafters-redis-go/  \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"init\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage #1: Bind to a port\"},{\"slug\":\"ping-pong\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage #2: Respond to PING\"},{\"slug\":\"ping-pong-multiple\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage #3: Respond to multiple PINGs\"},{\"slug\":\"concurrent-clients\",\"tester_log_prefix\":\"stage-4\",\"title\":\"Stage #4: Handle concurrent clients\"},{\"slug\":\"echo\",\"tester_log_prefix\":\"stage-5\",\"title\":\"Stage #5: Implement the ECHO command\"},{\"slug\":\"set_get\",\"tester_log_prefix\":\"stage-6\",\"title\":\"Stage #6: Implement the SET \u0026 GET commands\"},{\"slug\":\"expiry\",\"tester_log_prefix\":\"stage-7\",\"title\":\"Stage #7: Expiry\"},{\"slug\":\"repl-custom-port\",\"tester_log_prefix\":\"stage-101\",\"title\":\"Stage #101: Replication - Custom Port\"}, {\"slug\":\"repl-info\",\"tester_log_prefix\":\"stage-102\",\"title\":\"Stage #102: Replication - Info on Master\"}, {\"slug\":\"repl-info-replica\",\"tester_log_prefix\":\"stage-103\",\"title\":\"Stage #103: Replication - Info on Replica\"}, {\"slug\":\"repl-id\",\"tester_log_prefix\":\"stage-104\",\"title\":\"Stage #104: Replication - Replication ID and Offset\"}]" \
	$(shell pwd)/dist/main.out

copy_course_file:
	gh api repos/codecrafters-io/build-your-own-redis/contents/course-definition.yml \
		| jq -r .content \
		| base64 -d \
		> internal/test_helpers/course_definition.yml

record_fixtures:
	CODECRAFTERS_RECORD_FIXTURES=true make test

update_tester_utils:
	go get -u github.com/codecrafters-io/tester-utils
