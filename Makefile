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

test:
	go test -cover \
		github.com/athoune/redistop/circular
