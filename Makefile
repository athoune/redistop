GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)
DOCKER_GOLANG_VERSION=1.22-bookworm

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
		-d redis:8.6-alpine \
		    --requirepass test
	docker container list --filter 'name=redistop-test' --all

docker-redis-stop:
	docker container stop redistop-test
	docker container remove redistop-test

test-integration:
	make docker-redis-start
	go test -cover \
		github.com/athoune/redistop/monitor
	make docker-redis-stop

test:
	go test -cover \
		github.com/athoune/redistop/circular

test-all: test test-integration

redistop:
	REDISTOP_PASSWORD=test ./redistop
