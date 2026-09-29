package monitor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedis(t *testing.T) {
	r, err := Redis("127.0.0.1:6379", "test")
	require.NoError(t, err)
	require.NotNil(t, r.client)
	t.Cleanup(r.Close)

	s, err := r.Info()
	require.NoError(t, err)
	assert.NotEmpty(t, s)

	m, err := r.Memory()
	require.NoError(t, err)
	assert.NotNil(t, m)
}
