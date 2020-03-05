.PHONY: release build run run_with_redis

current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

release:
	git tag v$(next_version_number)
	git push origin master v$(next_version_number)

build:
	go build -o dist/main.out

make test:
	go test -v

test_with_redis: build
	CODECRAFTERS_SUBMISSION_DIR=./test_helpers/pass_all \
	CODECRAFTERS_CURRENT_STAGE_SLUG="expiry" \
	dist/main.out

test_tmp: build
	cd /tmp/45c297f9e27ea8dc && \
	CODECRAFTERS_SUBMISSION_DIR=/tmp/45c297f9e27ea8dc \
	CODECRAFTERS_CURRENT_STAGE_SLUG="ping-pong" \
	$(shell pwd)/dist/main.out

copy_course_file:
	hub api \
		repos/rohitpaulk/codecrafters-server/contents/codecrafters/store/data/redis.yml \
		| jq -r .content \
		| base64 -d \
		> test_helpers/course_definition.yml
