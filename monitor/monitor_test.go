package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMonitorLine(t *testing.T) {
	line, ok := parseMonitorLine("+1619454979.381488 [1 172.29.1.2:57676] \"brpop\"\n")
	require.True(t, ok)
	assert.Equal(t, "BRPOP", line.Command)
	assert.Equal(t, "172.29.1.2", line.IP)
	assert.Equal(t, 57676, line.port)
	assert.Equal(t, 1, line.n)

	line, ok = parseMonitorLine("+1621757323.274428 [9 [::1]:38824] \"DEL\" \"f00dc225-1975-4590-8a37-9b0ea4ec5acc\"\n")
	require.True(t, ok)
	assert.Equal(t, "DEL", line.Command)
	assert.Equal(t, "[::1]", line.IP)
	assert.Equal(t, 38824, line.port)

	line, ok = parseMonitorLine("+1621757323.279920 [9 lua] \"DEL\" \"f00dc225-1975-4590-8a37-9b0ea4ec5acc\"\n")
	require.True(t, ok)
	assert.Equal(t, "DEL", line.Command)
	assert.Equal(t, "lua", line.IP)
	assert.Equal(t, 0, line.port)

	for _, garbage := range []string{"+OK\n", "garbage\n", "", "+123 [x]\n"} {
		_, ok := parseMonitorLine(garbage)
		assert.False(t, ok, "input %q must not parse", garbage)
	}
}

// No server needed: a cancelled context must stop the loop promptly
// and close both channels instead of leaking the goroutine.
func TestMonitorCancelledContext(t *testing.T) {
	r := &RedisServer{address: "127.0.0.1:6379"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lines, errors := r.Monitor(ctx, func(bool) {})
	timeout := time.After(5 * time.Second)
	for lines != nil || errors != nil {
		select {
		case _, ok := <-lines:
			if !ok {
				lines = nil
			}
		case _, ok := <-errors:
			if !ok {
				errors = nil
			}
		case <-timeout:
			t.Fatal("Monitor did not exit on cancelled context")
		}
	}
}
