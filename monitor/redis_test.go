package monitor

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRedis(t *testing.T) {
	r, err := Redis("127.0.0.1:6379", "test")
	assert.NoError(t, err)
	assert.NotNil(t, r.client)
	var s map[string]string
	var m *MemoryStats
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		time.Sleep(10 * time.Millisecond)
		var err error
		s, err = r.Info()
		fmt.Println(s)
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	go func() {
		var err error
		m, err = r.Memory()
		if err != nil {
			panic(err)
		}
		fmt.Println(m)
		wg.Done()
	}()

	wg.Wait()
	assert.True(t, len(s) > 0)
	assert.NotNil(t, m)
}
