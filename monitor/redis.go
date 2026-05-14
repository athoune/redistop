package monitor

import (
	"context"
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
func Redis(address, password string) (*RedisServer, error) {
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

func (r *RedisServer) makeClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        r.address,
		Password:    r.password,
		DB:          0,
		PoolSize:    1,
		DialTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
