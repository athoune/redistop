package cli

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/athoune/redistop/circular"
	"github.com/athoune/redistop/stats"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *App) MonitorLoop(ctx context.Context) {
	statz := stats.New()
	var lock sync.Mutex

	lines, monitorErrors := a.redis.Monitor(ctx, func(ok bool) {
		if !ok {
			a.ui.Alert("Not connected")
		}
	})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-monitorErrors:
				a.ui.Alert(fmt.Sprintf("%v", err))
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-lines:
				if !ok {
					return
				}
				lock.Lock()
				statz.Feed(line)
				lock.Unlock()
			}
		}
	}()

	go func() {
		scale := float64(a.config.Frequency) / float64(time.Second)
		values := circular.NewCircular(250, scale)
		ticker := time.NewTicker(a.config.Frequency)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Snapshot under lock, render without holding it.
				lock.Lock()
				s := stats.Count(statz.Commands)
				ip := stats.Count(statz.Ips)
				statz.Reset()
				lock.Unlock()
				for _, i := range s {
					values.Add(i.V)
				}
				_, _, w, _ := a.ui.graph.GetInnerRect()
				size := w - 7
				if size < 1 {
					size = 1
				}
				vv := values.LastValues(size)
				var m float64
				for _, v := range vv {
					if v > m {
						m = v
					}
				}
				var current float64
				if len(vv) > 0 {
					current = vv[len(vv)-1]
				}
				values.Next()
				// All tview mutations happen on the UI goroutine.
				a.ui.app.QueueUpdateDraw(func() {
					if !a.ui.monitorIsReady {
						a.ui.monitorIsReady = true
						a.ui.grid.RemoveItem(a.ui.splash)
						a.ui.grid.AddItem(a.ui.cmds, 2, 0, 1, 1, 0, 0, false).
							AddItem(a.ui.ips, 2, 1, 1, 1, 0, 0, false)
					}
					a.ui.graph.SetSeries(vv)
					a.ui.graph.SetTitle(fmt.Sprintf("Commands [current: %.1f max: %.1f]",
						current,
						m,
					))
					a.ui.cmds.Clear()
					_, _, cw, _ := a.ui.cmds.GetInnerRect()
					if cw < 0 {
						cw = 0
					}
					for i, kv := range s {
						a.ui.cmds.SetCell(len(s)-i-1, 0,
							tview.NewTableCell(fmt.Sprintf("%-*s", cw/2, kv.K)).SetAttributes(tcell.AttrBold))
						a.ui.cmds.SetCell(len(s)-i-1, 1,
							tview.NewTableCell(fmt.Sprintf("%.1f", float64(kv.V)/scale)).
								SetAlign(tview.AlignRight))
					}

					a.ui.ips.Clear()
					_, _, iw, _ := a.ui.ips.GetInnerRect()
					if iw < 0 {
						iw = 0
					}
					for i, kv := range ip {
						a.ui.ips.SetCell(len(ip)-i-1, 0,
							tview.NewTableCell(fmt.Sprintf("%-*s", iw/2, kv.K)).SetAttributes(tcell.AttrItalic))
						a.ui.ips.SetCellSimple(len(ip)-i-1, 1, fmt.Sprintf("%.1f", float64(kv.V)/scale))
					}
				})
			}
		}
	}()
}
