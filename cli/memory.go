package cli

import (
	"fmt"
	"time"
)

func (a *App) MemoryLoop() {
	go func() {
		for {
			m, memErr := a.redis.Memory()
			if memErr != nil {
				a.log.Printf("Memory Error : %s", memErr.Error())
			}
			kv, infoErr := a.redis.Info()
			if infoErr != nil {
				a.log.Printf("Info Memory Error : %s", infoErr.Error())
			}
			// All tview mutations happen on the UI goroutine.
			// m may be nil when Memory() failed: never dereference it then.
			if memErr == nil || infoErr == nil {
				a.ui.app.QueueUpdateDraw(func() {
					if memErr == nil {
						a.ui.header.GetCell(0, 4).Text = fmt.Sprintf("keys: %d", m.KeysCount)
						a.ui.header.GetCell(0, 5).Text = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
					}
					if infoErr == nil {
						a.ui.memories.SetTitle(fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"]))
						if memErr == nil {
							for i, line := range m.Table() {
								for j, col := range line {
									a.ui.memories.GetCell(i, j).Text = col
								}
							}
						}
					}
				})
			}

			time.Sleep(5 * time.Second)
		}
	}()
}
