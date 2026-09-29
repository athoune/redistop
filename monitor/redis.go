package monitor

import (
	"context"
	"net"
	"time"

	"github.com/valkey-io/valkey-go"
)

// RedisServer talks to a Valkey (RESP compatible) server.
// MONITOR itself still uses a raw TCP connection (see monitor.go),
// this client is only used for INFO and MEMORY STATS.
type RedisServer struct {
	address  string
	password string
	client   valkey.Client
}

func Redis(address, password string) (*RedisServer, error) {
	r := &RedisServer{
		address:  address,
		password: password,
	}
	if err := r.makeClient(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *RedisServer) makeClient() error {
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:       []string{r.address},
		Password:          r.password,
		ForceSingleClient: true,
		Dialer:            net.Dialer{Timeout: 2 * time.Second},
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return err
	}
	r.client = client
	return nil
}

// Close releases the underlying connections.
func (r *RedisServer) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
