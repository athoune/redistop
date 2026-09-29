package stats

import (
	"testing"

	"github.com/athoune/redistop/monitor"
	"github.com/stretchr/testify/assert"
)

func TestFeedCountsCommandsAndIPs(t *testing.T) {
	s := New()
	s.Feed(monitor.Line{IP: "10.0.0.1", Command: "GET"})
	s.Feed(monitor.Line{IP: "10.0.0.1", Command: "GET"})
	s.Feed(monitor.Line{IP: "10.0.0.2", Command: "SET"})

	assert.Equal(t, map[string]int{"GET": 2, "SET": 1}, s.Commands)
	assert.Equal(t, map[string]int{"10.0.0.1": 2, "10.0.0.2": 1}, s.Ips)
}

func TestCountSortsByValueThenKey(t *testing.T) {
	got := Count(map[string]int{"B": 1, "A": 1, "C": 3, "D": 2})
	assert.Equal(t, ByValue{
		{K: "A", V: 1},
		{K: "B", V: 1},
		{K: "D", V: 2},
		{K: "C", V: 3},
	}, got)
}

func TestCountEmpty(t *testing.T) {
	assert.Empty(t, Count(map[string]int{}))
	assert.Empty(t, Count(nil))
}

func TestReset(t *testing.T) {
	s := New()
	s.Feed(monitor.Line{IP: "10.0.0.1", Command: "GET"})
	s.Reset()
	assert.Empty(t, s.Commands)
	assert.Empty(t, s.Ips)
	// Usable again after reset.
	s.Feed(monitor.Line{IP: "10.0.0.1", Command: "PING"})
	assert.Equal(t, map[string]int{"PING": 1}, s.Commands)
}
