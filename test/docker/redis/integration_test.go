// +build integration

package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	patronredis "github.com/beatlabs/patron/client/redis"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"

	"github.com/stretchr/testify/assert"
)

var (
	runtime *redisRuntime
)

func TestMain(m *testing.M) {
	var err error
	runtime, err = create(60 * time.Second)
	if err != nil {
		fmt.Printf("could not create redis runtime: %v\n", err)
		os.Exit(1)
	}
	defer func() {

	}()
	exitCode := m.Run()

	ee := runtime.Teardown()
	if len(ee) > 0 {
		for _, err = range ee {
			fmt.Printf("could not tear down containers: %v\n", err)
		}
	}
	os.Exit(exitCode)
}

func TestNew(t *testing.T) {
	cache, err := getRedisCache(runtime)
	assert.NoError(t, err)
	assert.NotNil(t, cache)
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
	addTestKey(runtime)

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, err1 := getRedisCache(runtime)
			assert.NoError(t, err1)
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

	cache, err1 := getRedisCache(runtime)
	assert.NoError(t, err1)
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
	addTestKey(runtime)
	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, err1 := getRedisCache(runtime)
			assert.NoError(t, err1)
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
	addTestKey(runtime)
	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			cache, err := getRedisCache(runtime)
			assert.NoError(t, err)
			val, exists, err1 := cache.Get(test.key)
			assert.NoError(t, err1)
			assert.Equal(t, test.wantvalue, val)
			assert.Equal(t, test.wantexists, exists)
		})
	}
}

func TestCache_SetTTL(t *testing.T) {

	cache, err1 := getRedisCache(runtime)
	assert.NoError(t, err1)
	err := cache.SetTTL("testTTL", "short", time.Second*10)
	assert.NoError(t, err)
	ttl := getTTL(runtime, "testTTL")
	assert.Equal(t, time.Second*10, ttl)

}

func TestClient_Ping(t *testing.T) {
	client := getRedisClient(runtime)
	mtr := mocktracer.New()
	opentracing.SetGlobalTracer(mtr)
	result := client.Ping(context.Background()).Val()
	assert.Equal(t, "PONG", result)
	assertSpan(t, mtr.FinishedSpans()[0])
}

func TestClient_Close(t *testing.T) {
	client := getRedisClient(runtime)
	mtr := mocktracer.New()
	opentracing.SetGlobalTracer(mtr)
	err := client.Close(context.Background())
	result := client.Ping(context.Background()).Val()
	assert.NoError(t, err)
	assert.Empty(t, result)
	assertSpan(t, mtr.FinishedSpans()[0])
}

func assertSpan(t *testing.T, sp *mocktracer.MockSpan) {
	assert.Equal(t, map[string]interface{}{
		"component":    patronredis.RedisComponent,
		"db.instance":  runtime.DSN(),
		"db.statement": "",
		"db.type":      patronredis.RedisDBType,
		"error":        false,
	}, sp.Tags())
}
