GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)
DOCKER_GOLANG_VERSION=1.27-alpine3.24
DOCKER_REDIS_VERSION=8-alpine
DOCKER_VALKEY_VERSION=9-alpine

build: bin
	go build \
		-o bin/redistop \
		-ldflags "-X github.com/athoune/redistop/version.version=$(GIT_VERSION)" \
		.

bin:
	mkdir -p bin

docker-build:
	mkdir -p .gocache
	docker run -t \
		-v `pwd`:/src \
		-v `pwd`/.gocache:/.cache \
		-v `pwd`/docker_gitconfig:/.gitconfig \
		-u `id -u` \
		-e GOCACHE=/.cache \
		-w /src \
		golang:${DOCKER_GOLANG_VERSION} \
		make
	[ -x "`which upx 2>/dev/null`" ] && upx bin/redistop
	file bin/redistop

docker-redis-start:
	docker run \
	    --name redistop-test \
		--publish 127.0.0.1:6379:6379 \
		-d redis:${DOCKER_REDIS_VERSION} \
		    --requirepass test
	docker container list --filter 'name=redistop-test' --all

docker-redis-stop:
	docker container stop redistop-test
	docker container remove redistop-test

docker-valkey-start:
	docker run \
	    --name redistop-valkey-test \
		--publish 127.0.0.1:6379:6379 \
		-d valkey/valkey:${DOCKER_VALKEY_VERSION} \
		    --requirepass test
	docker container list --filter 'name=redistop-valkey-test' --all

docker-valkey-stop:
	docker container stop redistop-valkey-test
	docker container remove redistop-valkey-test

test-redis-integration:
	make docker-redis-start
	go test -cover \
		github.com/athoune/redistop/monitor
	make docker-redis-stop

test-valkey-integration:
	make docker-valkey-start
	go test -cover \
		github.com/athoune/redistop/monitor
	make docker-valkey-stop

test-integration: test-valkey-integration test-redis-integration

test:
	go test -cover \
		github.com/athoune/redistop/circular

test-all: test test-integration

redistop:
	REDISTOP_PASSWORD=test ./redistop
