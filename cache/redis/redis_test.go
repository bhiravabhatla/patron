package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cache, err := New(context.Background(), Options{Addr: "localhost:6379"})
	assert.NoError(t, err)
	assert.NotNil(t, cache)

	want := "localhost:6379"
	got := cache.rdb.Options().Addr
	if want != got {
		t.Errorf("Wanted Address value %s; got %s", want, got)
	}
}

func TestCache_Get(t *testing.T) {
	tt := []struct {
		name       string
		key        string
		wantvalue  interface{}
		wantexists bool
		wanterror  bool
	}{
		{name: "Empty Key", key: "", wantvalue: nil, wantexists: false, wanterror: false},
		{name: "Existing Key", key: "test", wantvalue: "testvalue", wantexists: true, wanterror: false},
		{name: "Non Existing Key", key: "absent", wantvalue: nil, wantexists: false, wanterror: false},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, closefunc := setup(t)
			defer closefunc()
			val, exists, err := cache.Get(test.key)
			assert.NoError(t, err)
			assert.Equal(t, test.wantvalue, val)
			assert.Equal(t, test.wantexists, exists)
		})

	}

}

func TestCache_Set(t *testing.T) {
	tt := []struct {
		name      string
		key       string
		value     interface{}
		wanterror bool
	}{
		{name: "Empty Key", key: "", value: "test", wanterror: false},
		{name: "Empty Value", key: "empty", value: "", wanterror: false},
		{name: "Existing Key", key: "test", value: "newval", wanterror: false},
		{name: "Non Existing Key", key: "absent", value: "set", wanterror: false},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, closefunc := setup(t)
			defer closefunc()
			err := cache.Set(test.key, test.value)
			if test.wanterror {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				val, exists, _ := cache.Get(test.key)
				assert.True(t, exists)
				assert.Equal(t, test.value, val)
			}
		})

	}

}

func TestCache_Purge(t *testing.T) {

	cache, closefunc := setup(t)
	defer closefunc()
	err := cache.Purge()
	assert.NoError(t, err)
	val, exists, _ := cache.Get("test")
	assert.False(t, exists)
	assert.Nil(t, val)

}

func TestCache_Remove(t *testing.T) {
	tt := []struct {
		name    string
		key     string
		wanterr bool
	}{
		{name: "Remove Existing key", key: "test", wanterr: false},
		{name: "Remove non existing key", key: "nonexisting", wanterr: false},
		{name: "Remove empty key", key: "", wanterr: false},
	}
	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, closefunc := setup(t)
			defer closefunc()
			err := cache.Remove(test.key)
			if test.wanterr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				val, _, _ := cache.Get(test.key)
				assert.Nil(t, val)
			}

		})
	}
}

func TestCache_SetTTL(t *testing.T) {

	cache, closefunc := setup(t)
	defer closefunc()
	err := cache.SetTTL("testTTL", "short", time.Second*10)
	assert.NoError(t, err)
	ttl := cache.rdb.TTL("testTTL").Val()
	assert.Equal(t, time.Second*10, ttl)

}

func setup(t *testing.T) (cache *Cache, close func()) {
	t.Parallel()
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("setting up test redis server failed with : %v", err)
	}
	er := redisServer.Set("test", "testvalue")
	if er != nil {
		t.Fatalf("Adding test data to redis server failed with : %v", err)
	}

	cache, e := New(context.Background(), Options{
		Addr: redisServer.Addr()})

	if e != nil {
		t.Fatalf("Falied creating cache with : %v", err)
	}

	return cache, func() {
		redisServer.Close()
	}
}
