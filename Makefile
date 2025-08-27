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
	go test -count=1 -p 1 -v ./internal/...

copy_course_file:
	gh api repos/codecrafters-io/build-your-own-redis/contents/course-definition.yml \
		| jq -r .content \
		| base64 -d \
		> internal/test_helpers/course_definition.yml

record_fixtures:
	CODECRAFTERS_RECORD_FIXTURES=true make test

update_tester_utils:
	go get -u github.com/codecrafters-io/tester-utils

TEST_TARGET ?= test
RUNS ?= 100
test_flakiness:
	@$(foreach i,$(shell seq 1 $(RUNS)), \
		echo "Running iteration $(i)/$(RUNS) of \"make $(TEST_TARGET)\"" ; \
		make $(TEST_TARGET) > /tmp/test ; \
		if [ "$$?" -ne 0 ]; then \
			echo "Test failed on iteration $(i)" ; \
			cat /tmp/test ; \
			exit 1 ; \
		fi ;\
	)

test_base_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"jm1\",\"tester_log_prefix\":\"stage-1\",\"title\":\"Stage #1: Bind to a port\"},{\"slug\":\"rg2\",\"tester_log_prefix\":\"stage-2\",\"title\":\"Stage #2: Respond to PING\"},{\"slug\":\"wy1\",\"tester_log_prefix\":\"stage-3\",\"title\":\"Stage #3: Respond to multiple PINGs\"},{\"slug\":\"zu2\",\"tester_log_prefix\":\"stage-4\",\"title\":\"Stage #4: Handle concurrent clients\"},{\"slug\":\"qq0\",\"tester_log_prefix\":\"stage-5\",\"title\":\"Stage #5: Implement the ECHO command\"},{\"slug\":\"la7\",\"tester_log_prefix\":\"stage-6\",\"title\":\"Stage #6: Implement the SET \u0026 GET commands\"},{\"slug\":\"yz1\",\"tester_log_prefix\":\"stage-7\",\"title\":\"Stage #7: Expiry\"}]" \
	dist/main.out

test_repl_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"bw1\",\"tester_log_prefix\":\"stage-101\",\"title\":\"Stage #101: Replication - Custom Port\"}, {\"slug\":\"ye5\",\"tester_log_prefix\":\"stage-102\",\"title\":\"Stage #102: Replication - Info on Master\"},{\"slug\":\"hc6\",\"tester_log_prefix\":\"stage-103\",\"title\":\"Stage #103: Replication - Info on Replica\"}, {\"slug\":\"xc1\",\"tester_log_prefix\":\"stage-104\",\"title\":\"Stage #104: Replication - Replication ID and Offset\"}, {\"slug\":\"gl7\",\"tester_log_prefix\":\"stage-105\",\"title\":\"Stage #105: Replication - Handshake 1\"},{\"slug\":\"eh4\",\"tester_log_prefix\":\"stage-106\",\"title\":\"Stage #106: Replication - Handshake 2\"},{\"slug\":\"ju6\",\"tester_log_prefix\":\"stage-107\",\"title\":\"Stage #107: Replication - Handshake 3\"},{\"slug\":\"fj0\",\"tester_log_prefix\":\"stage-108\",\"title\":\"Stage #108: Replication - REPLCONF\"},{\"slug\":\"vm3\",\"tester_log_prefix\":\"stage-109\",\"title\":\"Stage #109: Replication - PSYNC\"},{\"slug\":\"cf8\",\"tester_log_prefix\":\"stage-110\",\"title\":\"Stage #110: Replication - PSYNC w RDB file\"},{\"slug\":\"zn8\",\"tester_log_prefix\":\"stage-111\",\"title\":\"Stage #111: Command Propagation\"},{\"slug\":\"hd5\",\"tester_log_prefix\":\"stage-112\",\"title\":\"Stage #112: Command Propagation to multiple Replicas\"},{\"slug\":\"yg4\",\"tester_log_prefix\":\"stage-113\",\"title\":\"Stage #113: Command Processing\"},{\"slug\":\"xv6\",\"tester_log_prefix\":\"stage-114\",\"title\":\"Stage #114: GetAck with 0 offset\"},{\"slug\":\"yd3\",\"tester_log_prefix\":\"stage-115\",\"title\":\"Stage #115: GetAck with non-0 offset\"},{\"slug\":\"my8\",\"tester_log_prefix\":\"stage-116\",\"title\":\"Stage #116: WAIT with 0 replicas\"},{\"slug\":\"tu8\",\"tester_log_prefix\":\"stage-117\",\"title\":\"Stage #117: WAIT with 0 offset\"},{\"slug\":\"na2\",\"tester_log_prefix\":\"stage-118\",\"title\":\"Stage #118: WAIT Command\"}]" \
	dist/main.out

test_rdb_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"zg5\",\"tester_log_prefix\":\"stage-201\",\"title\":\"Stage #1: RDB Config\"}, {\"slug\":\"jz6\",\"tester_log_prefix\":\"stage-202\",\"title\":\"Stage #2: RDB Read Key\"}, {\"slug\":\"gc6\",\"tester_log_prefix\":\"stage-203\",\"title\":\"Stage #3: RDB String Value\"}, {\"slug\":\"jw4\",\"tester_log_prefix\":\"stage-204\",\"title\":\"Stage #4: RDB Read Multiple Keys\"}, {\"slug\":\"dq3\",\"tester_log_prefix\":\"stage-205\",\"title\":\"Stage #5: RDB Read Multiple String Values\"}, {\"slug\":\"sm4\",\"tester_log_prefix\":\"stage-206\",\"title\":\"Stage #6: RDB Read Value With Expiry\"}]" \
	dist/main.out

test_streams_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\": \"cc3\", \"tester_log_prefix\": \"stage-301\", \"title\": \"stage #01: StreamsType\"},{\"slug\": \"cf6\", \"tester_log_prefix\": \"stage-302\", \"title\": \"stage #02: StreamsXadd\"},{\"slug\": \"hq8\", \"tester_log_prefix\": \"stage-303\", \"title\": \"stage #03: StreamsXaddValidateID\"},{\"slug\": \"yh3\", \"tester_log_prefix\": \"stage-304\", \"title\": \"stage #04: StreamsXaddPartialAutoid\"},{\"slug\": \"xu6\", \"tester_log_prefix\": \"stage-305\", \"title\": \"stage #05: StreamsXaddFullAutoid\"},{\"slug\": \"zx1\", \"tester_log_prefix\": \"stage-306\", \"title\": \"stage #06: StreamsXrange\"},{\"slug\": \"yp1\", \"tester_log_prefix\": \"stage-307\", \"title\": \"stage #07: StreamsXrangeMinID\"},{\"slug\": \"fs1\", \"tester_log_prefix\": \"stage-308\", \"title\": \"stage #08: StreamsXrangeMaxID\"},{\"slug\": \"um0\", \"tester_log_prefix\": \"stage-309\", \"title\": \"stage #09: StreamsXread\"},{\"slug\": \"ru9\", \"tester_log_prefix\": \"stage-310\", \"title\": \"stage #10: StreamsXreadMultiple\"},{\"slug\": \"bs1\", \"tester_log_prefix\": \"stage-311\", \"title\": \"stage #11: StreamsXreadBlock\"},{\"slug\": \"hw1\", \"tester_log_prefix\": \"stage-312\", \"title\": \"stage #12: StreamsXreadBlockNoTimeout\"},{\"slug\": \"xu1\", \"tester_log_prefix\": \"stage-313\", \"title\": \"stage #13: StreamsXreadBlockMaxID\"}]"  \
	dist/main.out

test_txn_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"si4\",\"tester_log_prefix\":\"stage-401\",\"title\":\"Stage #401: INCR-1\"},{\"slug\":\"lz8\",\"tester_log_prefix\":\"stage-402\",\"title\":\"Stage #402: INCR-2\"}, {\"slug\":\"mk1\",\"tester_log_prefix\":\"stage-403\",\"title\":\"Stage #403: INCR-3\"}, {\"slug\":\"pn0\",\"tester_log_prefix\":\"stage-404\",\"title\":\"Stage #404: MULTI\"}, {\"slug\":\"lo4\",\"tester_log_prefix\":\"stage-405\",\"title\":\"Stage #405: EXEC\"}, {\"slug\":\"we1\",\"tester_log_prefix\":\"stage-406\",\"title\":\"Stage #406: Empty Transaction\"}, {\"slug\":\"rs9\",\"tester_log_prefix\":\"stage-407\",\"title\":\"Stage #407: Queueing Commands\"}, {\"slug\":\"fy6\",\"tester_log_prefix\":\"stage-408\",\"title\":\"Stage #408: Executing a transaction\"}, {\"slug\":\"rl9\",\"tester_log_prefix\":\"stage-409\",\"title\":\"Stage #409: Discarding a transaction\"}, {\"slug\":\"sg9\",\"tester_log_prefix\":\"stage-410\",\"title\":\"Stage #410: Executing a failed transaction\"}, {\"slug\":\"jf8\",\"tester_log_prefix\":\"stage-411\",\"title\":\"Stage #411: Executing concurrent transactions\"}]" \
	dist/main.out

test_list_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"mh6\",\"tester_log_prefix\":\"stage-501\",\"title\":\"Stage #501: RPUSH-1\"},{\"slug\":\"tn7\",\"tester_log_prefix\":\"stage-502\",\"title\":\"Stage #502: RPUSH-2\"}, {\"slug\":\"lx4\",\"tester_log_prefix\":\"stage-503\",\"title\":\"Stage #503: RPUSH-3\"}, {\"slug\":\"sf6\",\"tester_log_prefix\":\"stage-504\",\"title\":\"Stage #504: LRANGE-1\"}, {\"slug\":\"ri1\",\"tester_log_prefix\":\"stage-505\",\"title\":\"Stage #505: LRANGE-2\"}, {\"slug\":\"gu5\",\"tester_log_prefix\":\"stage-506\",\"title\":\"Stage #506: LPUSH\"}, {\"slug\":\"fv6\",\"tester_log_prefix\":\"stage-507\",\"title\":\"Stage #507: LLEN\"},{\"slug\":\"ef1\",\"tester_log_prefix\":\"stage-508\",\"title\":\"Stage #508: LPOP-1\"}, {\"slug\":\"jp1\",\"tester_log_prefix\":\"stage-509\",\"title\":\"Stage #509: LPOP-2\"}, {\"slug\":\"ec3\",\"tester_log_prefix\":\"stage-510\",\"title\":\"Stage #510: BLPOP-1\"}, {\"slug\":\"xj7\",\"tester_log_prefix\":\"stage-511\",\"title\":\"Stage #511: BLPOP-2\"}]" \
	dist/main.out

test_pubsub_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"mx3\",\"tester_log_prefix\":\"stage-601\",\"title\":\"Stage #601: SUBSCRIBE-1\"},{\"slug\":\"zc8\",\"tester_log_prefix\":\"stage-602\",\"title\":\"Stage #602: SUBSCRIBE-2\"}, {\"slug\":\"aw8\",\"tester_log_prefix\":\"stage-603\",\"title\":\"Stage #603: SUBSCRIBE-3\"}, {\"slug\":\"lf1\",\"tester_log_prefix\":\"stage-604\",\"title\":\"Stage #604: SUBSCRIBE-4\"}, {\"slug\":\"hf2\",\"tester_log_prefix\":\"stage-605\",\"title\":\"Stage #605: PUBLISH-1\"}, {\"slug\":\"dn4\",\"tester_log_prefix\":\"stage-606\",\"title\":\"Stage #606: PUBLISH-2\"}, {\"slug\":\"ze9\",\"tester_log_prefix\":\"stage-607\",\"title\":\"Stage #607: UNSUBSCRIBE\"}]" \
	dist/main.out

test_zset_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"ct1\",\"tester_log_prefix\":\"stage-701\",\"title\":\"Stage #701: ZADD-1\"},{\"slug\":\"hf1\",\"tester_log_prefix\":\"stage-702\",\"title\":\"Stage #702: ZADD-2\"}, {\"slug\":\"lg6\",\"tester_log_prefix\":\"stage-703\",\"title\":\"Stage #703: ZRANK\"}, {\"slug\":\"ic1\",\"tester_log_prefix\":\"stage-704\",\"title\":\"Stage #704: ZRANGE-1\"}, {\"slug\":\"bj4\",\"tester_log_prefix\":\"stage-705\",\"title\":\"Stage #705: ZRANGE-2\"}, {\"slug\":\"kn4\",\"tester_log_prefix\":\"stage-706\",\"title\":\"Stage #706: ZCARD\"}, {\"slug\":\"gd7\",\"tester_log_prefix\":\"stage-707\",\"title\":\"Stage #707: ZSCORE\"}, {\"slug\":\"sq7\",\"tester_log_prefix\":\"stage-708\",\"title\":\"Stage #708: ZREM\"} ]" \
	dist/main.out

test_geospatial_with_redis: build
	CODECRAFTERS_REPOSITORY_DIR=./internal/test_helpers/pass_all \
	CODECRAFTERS_TEST_CASES_JSON="[{\"slug\":\"zt4\",\"tester_log_prefix\":\"stage-801\",\"title\":\"Stage #801: GEOADD-1\"},{\"slug\":\"ck3\",\"tester_log_prefix\":\"stage-802\",\"title\":\"Stage #802: GEOADD-2\"}, {\"slug\":\"tn5\",\"tester_log_prefix\":\"stage-803\",\"title\":\"Stage #803: GEOADD-3\"}, {\"slug\":\"cr3\",\"tester_log_prefix\":\"stage-804\",\"title\":\"Stage #804: GEOADD-4\"}, {\"slug\":\"xg4\",\"tester_log_prefix\":\"stage-805\",\"title\":\"Stage #805: GEOPOS-1\"}, {\"slug\":\"hb5\",\"tester_log_prefix\":\"stage-806\",\"title\":\"Stage #806: GEOPOS-2\"}, {\"slug\":\"ek6\",\"tester_log_prefix\":\"stage-807\",\"title\":\"Stage #807: GEODIST\"}, {\"slug\":\"rm9\",\"tester_log_prefix\":\"stage-808\",\"title\":\"Stage #808: GEOSEARCH\"}]" \
	dist/main.out

test_all_with_redis:
	make test_base_with_redis || true
	make test_repl_with_redis || true
	make test_rdb_with_redis || true
	make test_streams_with_redis || true
	make test_txn_with_redis || true
	make test_list_with_redis || true
	make test_pubsub_with_redis || true
	make test_zset_with_redis || true
	make test_geospatial_with_redis || true

setup:
	echo "Setting up redis-tester prerequisites for Linux"

	curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg
	sudo chmod 644 /usr/share/keyrings/redis-archive-keyring.gpg
	@echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(shell lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

	sudo apt-get update && sudo apt-get install redis -y

	sudo service redis-server stop

	echo "Setup complete!"