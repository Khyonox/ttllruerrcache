package ttllruerrcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	var c Cache
	_, exists := c.Get("hello")
	require.False(t, exists)
	c.Set("hello", "world")

	val, _ := c.Get("hello")
	require.Equal(t, val, "world")

	now := time.Now()
	c.SetFull("name", "bob", now, time.Second)
	val, _ = c.GetFull("name", now)
	require.Equal(t, val, "bob")

	val, _ = c.GetFull("name", now.Add(time.Second*2))
	require.Nil(t, val)

	val, _ = c.GetFull("name", now)
	require.Nil(t, val)
}
