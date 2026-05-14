package monitor

import "context"

// Info fetches the INFO command output and parses it into a key/value map.
func (r *RedisServer) Info() (map[string]string, error) {
	ctx := context.Background()
	bulk, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}
	return BulkTable(bulk)
}
