package monitor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

type MemoryStats struct {
	PeakAllocated      int64
	DatasetBytes       int64
	KeysCount          int64
	Fragmentation      float64
	ReplicationBacklog int64
}

func (r *RedisServer) Memory() (*MemoryStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res := r.client.Do(ctx, r.client.B().Arbitrary("MEMORY", "STATS").Build())
	if err := res.Error(); err != nil {
		return nil, err
	}
	m := &MemoryStats{}
	// RESP3 servers may return a map, RESP2 a flat array [k, v, k, v...].
	// Try map first, fall back to array.
	if asMap, err := res.AsMap(); err == nil {
		for k, v := range asMap {
			if err := m.feed(k, v); err != nil {
				return nil, err
			}
		}
		return m, nil
	}
	arr, err := res.ToArray()
	if err != nil {
		return nil, fmt.Errorf("memory stats: unexpected response type: %w", err)
	}
	if len(arr)%2 != 0 {
		return nil, fmt.Errorf("memory stats: odd array length %d", len(arr))
	}
	for i := 0; i < len(arr); i += 2 {
		key, err := arr[i].ToString()
		if err != nil {
			return nil, fmt.Errorf("memory stats key: %w", err)
		}
		if err := m.feed(key, arr[i+1]); err != nil {
			return nil, err
		}
	}
	return m, nil
}

func (m *MemoryStats) feed(key string, v valkey.ValkeyMessage) error {
	switch key {
	case "peak.allocated":
		vv, err := messageToInt(v)
		if err != nil {
			return fmt.Errorf("peak.allocated: %w", err)
		}
		m.PeakAllocated = vv
	case "dataset.bytes":
		vv, err := messageToInt(v)
		if err != nil {
			return fmt.Errorf("dataset.bytes: %w", err)
		}
		m.DatasetBytes = vv
	case "keys.count":
		vv, err := messageToInt(v)
		if err != nil {
			return fmt.Errorf("keys.count: %w", err)
		}
		m.KeysCount = vv
	case "fragmentation":
		vv, err := messageToFloat(v)
		if err != nil {
			return fmt.Errorf("fragmentation: %w", err)
		}
		m.Fragmentation = vv
	case "replication.backlog":
		vv, err := messageToInt(v)
		if err != nil {
			return fmt.Errorf("replication.backlog: %w", err)
		}
		m.ReplicationBacklog = vv
	}
	return nil
}

func messageToInt(v valkey.ValkeyMessage) (int64, error) {
	if i, err := v.AsInt64(); err == nil {
		return i, nil
	}
	s, err := v.ToString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}

func messageToFloat(v valkey.ValkeyMessage) (float64, error) {
	if f, err := v.AsFloat64(); err == nil {
		return f, nil
	}
	s, err := v.ToString()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}

func (m *MemoryStats) Table() [][]string {
	return [][]string{
		{"peak allocated", fmt.Sprintf("%d", m.PeakAllocated)},
		{"dataset", fmt.Sprintf("%d bytes", m.DatasetBytes)},
		{"fragmentation", fmt.Sprintf("%.2f", m.Fragmentation)},
		{"repl.backlog", fmt.Sprintf("%d", m.ReplicationBacklog)},
	}
}
