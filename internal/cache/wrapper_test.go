package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {

	type tStruct struct {
		Param1 int64
		Param2 string
	}

	var (
		ctx            = context.Background()
		key            = "key:1"
		expectedStruct = tStruct{Param1: int64(10), Param2: "movieID"}
	)
	t.Run("Cache sets value if it does not exist", func(t *testing.T) {
		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &Cache{Client: c}

		v, _ := rs.Get(key)

		_, err := GetOrSet(cache, ctx, key, time.Minute, func() (tStruct, error) {
			return expectedStruct, nil
		})

		savedString, _ := rs.Get(key)
		expBytes, _ := json.Marshal(expectedStruct)

		require.NoError(t, err)
		require.Equal(t, "", v)
		require.Equal(t, string(expBytes), savedString)

	})

	t.Run("Cache returns cached struct", func(t *testing.T) {
		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &Cache{Client: c}
		expBytes, _ := json.Marshal(expectedStruct)
		rs.Set(key, string(expBytes))

		s, err := GetOrSet(cache, ctx, key, time.Minute, func() (tStruct, error) {
			return tStruct{}, nil
		})

		require.NoError(t, err)
		require.Equal(t, expectedStruct, s)

	})

	t.Run("Cache returns error for invalid stored data", func(t *testing.T) {
		rs := miniredis.RunT(t)
		c := redis.NewClient(&redis.Options{Addr: rs.Addr()})
		cache := &Cache{Client: c}
		_ = rs.Set(key, "invalid_data")

		s, err := GetOrSet(cache, ctx, key, time.Minute, func() (tStruct, error) {
			return tStruct{}, nil
		})

		require.IsType(t, err, &json.SyntaxError{})
		require.Equal(t, tStruct{}, s)

	})

}
