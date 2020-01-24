.PHONY: release build run run_with_redis

current_version_number := $(shell git tag --list "v*" | sort -V | tail -n 1 | cut -c 2-)
next_version_number := $(shell echo $$(($(current_version_number)+1)))

release_docker:
	git push origin master
	git tag v$(next_version_number)
	git push origin v$(next_version_number)

release:
	rm -rf dist
	goreleaser

build:
	go build -o dist/main.out

make test:
	go test -v

test_with_redis: build
	APP_DIR=./test_helpers/pass_all dist/main.out

bump_version:
	bumpversion --verbose --tag patch

bump_release_upload: bump_version release upload_to_travis
