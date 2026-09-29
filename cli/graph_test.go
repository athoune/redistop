package cli

import (
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/require"
)

// Draw on one goroutine while SetSeries hammers from another.
// Without the SetSeries lock, -race flags the series header.
// Run with: go test -race ./cli/
func TestGraphBoxConcurrentSetSeries(t *testing.T) {
	g := NewGraphBox()
	g.SetRect(0, 0, 80, 24)
	screen := tcell.NewSimulationScreen("")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	screen.SetSize(80, 24)

	done := make(chan struct{})
	time.AfterFunc(200*time.Millisecond, func() { close(done) })
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				g.Draw(screen)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
				g.SetSeries([]float64{float64(i), float64(i + 1), float64(i + 2)})
			}
		}
	}()
	wg.Wait()
}

func TestGraphBoxDrawDegenerateRect(t *testing.T) {
	g := NewGraphBox()
	g.SetSeries([]float64{1, 2, 3})
	screen := tcell.NewSimulationScreen("")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	// Too small for an inner rect: must not panic.
	g.SetRect(0, 0, 1, 1)
	screen.SetSize(1, 1)
	require.NotPanics(t, func() { g.Draw(screen) })
}
