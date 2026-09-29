package cli

import (
	"sync"
	"testing"
	"time"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Inline runner: no tview app running, so QueueUpdateDraw would block.
// The mutex models the production invariant (every TextView access
// happens on the single UI goroutine): update and reads share it,
// otherwise the test harness itself would race with the clear timer.
// Run with: go test -race ./cli/
func newTestLogger(ttl time.Duration) (*Logger, func() string) {
	tv := tview.NewTextView()
	var mu sync.Mutex
	get := func() string {
		mu.Lock()
		defer mu.Unlock()
		return tv.GetText(false)
	}
	l := &Logger{
		block: tv,
		update: func(f func()) {
			mu.Lock()
			defer mu.Unlock()
			f()
		},
		ttl: ttl,
	}
	return l, get
}

func TestLoggerConcurrentPrintf(t *testing.T) {
	l, get := newTestLogger(time.Minute)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l.Printf("msg %d", i)
		}(i)
	}
	wg.Wait()
	text := get()
	assert.Contains(t, text, "msg ")
	assert.NotEmpty(t, text)
}

func TestLoggerClearsAfterTTL(t *testing.T) {
	l, get := newTestLogger(50 * time.Millisecond)
	l.Printf("boom")
	require.Equal(t, "boom", get())
	time.Sleep(300 * time.Millisecond)
	assert.Empty(t, get())
}

func TestLoggerLastMessageWins(t *testing.T) {
	l, get := newTestLogger(80 * time.Millisecond)
	l.Printf("first")
	time.Sleep(40 * time.Millisecond)
	l.Printf("second")
	time.Sleep(200 * time.Millisecond)
	assert.Empty(t, get())
	// A fresh message after expiry still shows.
	l.Printf("third %d", 3)
	assert.Equal(t, "third 3", get())
}
