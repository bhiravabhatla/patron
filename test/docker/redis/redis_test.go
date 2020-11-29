package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"

	patronredis "github.com/beatlabs/patron/client/redis"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"

	rediscache "github.com/beatlabs/patron/cache/redis"
	patronDocker "github.com/beatlabs/patron/test/docker"
)

const (
	connectionFormat = "localhost:%s"
)

// redisRuntime defines a docker Redis runtime.
type redisRuntime struct {
	patronDocker.Runtime
}

// Create initializes a redis docker runtime.
func create(expiration time.Duration) (*redisRuntime, error) {
	br, err := patronDocker.NewRuntime(expiration)
	if err != nil {
		return nil, fmt.Errorf("could not create base runtime: %w", err)
	}

	runtime := &redisRuntime{Runtime: *br}

	runOptions := &dockertest.RunOptions{Repository: "redis",
		Tag: "6.0",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"6379/tcp": {{HostIP: "", HostPort: ""}},
		},
		//ExposedPorts: []string{"6379/tcp"},
	}
	_, err = runtime.RunWithOptions(runOptions)
	if err != nil {
		return nil, fmt.Errorf("could not start redis: %w", err)
	}
	err = runtime.Pool().Retry(func() error {
		db := redis.NewClient(&redis.Options{Addr: runtime.DSN()})
		return db.Ping().Err()
	})

	if err != nil {
		for _, err1 := range runtime.Teardown() {
			fmt.Printf("failed to teardown: %v\n", err1)
		}
		return nil, fmt.Errorf("container not ready: %w", err)
	}

	return runtime, nil
}

// Port returns a port where the container service can be reached.
func (s *redisRuntime) Port() string {
	return s.Resources()[0].GetPort("6379/tcp")
}

// DSN of the runtime.
func (s *redisRuntime) DSN() string {
	return fmt.Sprintf(connectionFormat, s.Port())
}

func getRedisCache(r *redisRuntime) (*rediscache.Cache, error) {
	return rediscache.New(context.Background(), rediscache.Options{Addr: r.DSN()})
}

func getRedisClient(r *redisRuntime) *patronredis.Client {
	return patronredis.New(patronredis.Options{Addr: r.DSN()})
}

func addTestKey(r *redisRuntime) {
	client := getRedisClient(r)
	client.Set("test", "testvalue", 0)
}

func getTTL(r *redisRuntime, key string) time.Duration {
	client := getRedisClient(r)
	return client.TTL(key).Val()
}
