package cli

import (
	"fmt"
	"sync"
	"time"

	"github.com/rivo/tview"
)

// Logger shows transient messages in a TextView, clearing them
// ttl after the last one. It is safe for concurrent use.
//
// update posts a func on the UI goroutine (tview widgets must only
// be touched there). Production passes app.QueueUpdateDraw; tests
// pass an inline runner. Nil update runs inline.
type Logger struct {
	block  *tview.TextView
	update func(func())
	ttl    time.Duration

	mu    sync.Mutex
	timer *time.Timer
	gen   uint64
}

const defaultLoggerTTL = 5 * time.Second

func (l *Logger) Printf(tpl string, args ...any) {
	msg := fmt.Sprintf(tpl, args...)
	update := l.update
	if update == nil {
		update = func(f func()) { f() }
	}
	ttl := l.ttl
	if ttl <= 0 {
		ttl = defaultLoggerTTL
	}
	l.mu.Lock()
	if l.timer != nil {
		l.timer.Stop()
	}
	l.gen++
	gen := l.gen
	l.timer = time.AfterFunc(ttl, func() {
		l.mu.Lock()
		stale := gen != l.gen
		l.mu.Unlock()
		if stale {
			return
		}
		update(func() { l.block.SetText("") })
	})
	l.mu.Unlock()
	update(func() { l.block.SetText(msg) })
}
