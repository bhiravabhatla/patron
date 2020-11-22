package redis

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/beatlabs/patron/trace"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpan(t *testing.T) {
	opts := Options{Addr: "localhost"}
	c := New(opts)
	mtr := mocktracer.New()
	opentracing.SetGlobalTracer(mtr)
	tag := opentracing.Tag{Key: "key", Value: "value"}
	sp, req := c.startSpan(context.Background(), "localhost", "flushdb", tag)
	assert.NotNil(t, sp)
	assert.NotNil(t, req)
	assert.IsType(t, &mocktracer.MockSpan{}, sp)
	jsp := sp.(*mocktracer.MockSpan)
	assert.NotNil(t, jsp)
	trace.SpanSuccess(sp)
	rawSpan := mtr.FinishedSpans()[0]
	assert.Equal(t, map[string]interface{}{
		"component":    RedisComponent,
		"db.instance":  "localhost",
		"db.statement": "flushdb",
		"db.type":      RedisDBType,
		"error":        false,
		"key":          "value",
	}, rawSpan.Tags())
}

func TestClient_Ping(t *testing.T) {
	client, closefunc := setup(t)
	defer closefunc()
	mtr := mocktracer.New()
	opentracing.SetGlobalTracer(mtr)
	result := client.Ping(context.Background()).Val()
	rawSpan := mtr.FinishedSpans()[0]
	assert.Equal(t, "PONG", result)
	assert.Equal(t, map[string]interface{}{
		"component":    RedisComponent,
		"db.instance":  client.Options().Addr,
		"db.statement": "",
		"db.type":      RedisDBType,
		"error":        false,
	}, rawSpan.Tags())

}

func TestClient_Close(t *testing.T) {
	client, closefunc := setup(t)
	defer closefunc()
	mtr := mocktracer.New()
	opentracing.SetGlobalTracer(mtr)
	err := client.Close(context.Background())
	result := client.Ping(context.Background()).Val()
	rawSpan := mtr.FinishedSpans()[0]
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.Equal(t, map[string]interface{}{
		"component":    RedisComponent,
		"db.instance":  client.Options().Addr,
		"db.statement": "",
		"db.type":      RedisDBType,
		"error":        false,
	}, rawSpan.Tags())
}

func setup(t *testing.T) (*Client, func()) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to set up mock redis server with %v", err)
	}
	return New(Options{Addr: s.Addr()}), func() {
		s.Close()
	}
}
