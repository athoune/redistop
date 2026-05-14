package monitor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisServer wraps a go-redis client for a single Redis instance.
type RedisServer struct {
	address  string
	password string
	client   *redis.Client
}

// Redis creates a new RedisServer connected to the given address.
// It pings the server to verify connectivity before returning.
// If password is empty, it tries to read REDISCLI_AUTH env var or ~/.rediscli_auth.
func Redis(address, password string) (*RedisServer, error) {
	if password == "" {
		password = resolvePassword()
	}
	r := &RedisServer{
		address:  address,
		password: password,
	}
	var err error
	r.client, err = r.makeClient()
	if err != nil {
		return nil, err
	}
	return r, nil
}

// resolvePassword tries to find a Redis password from common sources.
func resolvePassword() string {
	// 1. Environment variable used by redis-cli
	if pw := os.Getenv("REDISCLI_AUTH"); pw != "" {
		return pw
	}

	// 2. redis-cli auth file
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".rediscli_auth")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func (r *RedisServer) makeClient() (*redis.Client, error) {
	opts := &redis.Options{
		Addr:        r.address,
		DB:          0,
		PoolSize:    1,
		DialTimeout: 2 * time.Second,
		// Force RESP2 to avoid the HELLO 3 command which requires
		// authentication even for the handshake on password-protected servers.
		// This matches redis-cli default behaviour.
		Protocol: 2,
	}
	if r.password != "" {
		opts.Password = r.password
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
