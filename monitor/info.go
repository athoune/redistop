package monitor

import (
	"context"
	"time"
)

func (r *RedisServer) Info() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	bulk, err := r.client.Do(ctx, r.client.B().Info().Build()).ToString()
	if err != nil {
		return nil, err
	}
	return BulkTable(bulk)
}
