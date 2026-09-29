Redis Top
=========

Redistop uses [MONITOR](https://valkey.io/commands/monitor/) to watch Valkey (and Redis)
commands and shows per command and per host statistics.

> Because MONITOR streams back all commands, its use comes at a cost.

Redistop uses [INFO](https://valkey.io/commands/info/) command too.

"Redis stop" means "Say stop again" in french, and it's hard to find a new joke with Valkey.

Example
-------

![Redis Top screenshot](redistop.png)

Build
-----

Requires Go 1.26 or newer. If you have a Go dev environment set, you can build it with the Makefile

    make

If you need a Linux compilation, or just using Docker:

    make docker-build

Usage
-----

    ./bin/redistop [host:port] [password]

The default host is `localhost:6379`. The password can also come from the
`REDISTOP_PASSWORD` environment variable, which takes precedence over the
command line argument.

Tests
-----

Unit tests, no server needed:

    make test

Integration tests, need Docker, run against Valkey then Redis:

    make test-integration

License
-------

GPL v3
