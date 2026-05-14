package monitor

import (
	"context"
	"fmt"
	"strconv"
)

// MemoryStats holds a subset of fields returned by the MEMORY STATS command.
type MemoryStats struct {
	PeakAllocated      int64
	DatasetBytes       int64
	KeysCount          int64
	Fragmentation      float64
	ReplicationBacklog int64
}

// Memory queries Redis for MEMORY STATS and extracts the fields we care about.
func (r *RedisServer) Memory() (*MemoryStats, error) {
	m := &MemoryStats{}
	ctx := context.Background()

	cmd := r.client.Do(ctx, "MEMORY", "STATS")
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	// MEMORY STATS returns an array of alternating keys and values.
	raw, err := cmd.Result()
	if err != nil {
		return nil, err
	}

	// go-redis may parse the response as either []interface{} or map[interface{}]interface{}.
	mem := make(map[string]interface{})
	switch val := raw.(type) {
	case []interface{}:
		for i := 0; i+1 < len(val); i += 2 {
			key, ok := val[i].(string)
			if !ok {
				continue
			}
			mem[key] = val[i+1]
		}
	case map[interface{}]interface{}:
		for k, v := range val {
			key, ok := k.(string)
			if !ok {
				continue
			}
			mem[key] = v
		}
	default:
		return nil, fmt.Errorf("unexpected MEMORY STATS response type: %T", raw)
	}

	for k, v := range mem {
		switch k {
		case "peak.allocated":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.PeakAllocated = vv
		case "dataset.bytes":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.DatasetBytes = vv
		case "keys.count":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.KeysCount = vv
		case "fragmentation":
			// go-redis may return the value as float64, []byte or string
			// depending on the Redis version and driver internals.
			switch vv := v.(type) {
			case float64:
				m.Fragmentation = vv
			case []byte:
				vvv, err := strconv.ParseFloat(string(vv), 64)
				if err != nil {
					return nil, err
				}
				m.Fragmentation = vvv
			case string:
				vvv, err := strconv.ParseFloat(vv, 64)
				if err != nil {
					return nil, err
				}
				m.Fragmentation = vvv
			default:
				return nil, fmt.Errorf("unexpected fragmentation type: %T", v)
			}
		case "replication.backlog":
			vv, ok := v.(int64)
			if !ok {
				return nil, fmt.Errorf("not an int : %v", v)
			}
			m.ReplicationBacklog = vv
		}
	}
	return m, nil
}

// Table returns a human-readable representation of MemoryStats.
func (m *MemoryStats) Table() [][]string {
	return [][]string{
		{"peak allocated", fmt.Sprintf("%d", m.PeakAllocated)},
		{"dataset", fmt.Sprintf("%d bytes", m.DatasetBytes)},
		{"fragmentation", fmt.Sprintf("%.2f", m.Fragmentation)},
		{"repl.backlog", fmt.Sprintf("%d", m.ReplicationBacklog)},
	}
}
