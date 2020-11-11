package config

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func SetupRedis(ctx context.Context, env Env) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		DB:       0,
		Addr:     fmt.Sprintf("%s:%d", env.RedisHost, env.RedisPort),
		Password: env.RedisPassword,
	})

	// test the connection
	_, err := client.Ping(ctx).Result()

	return client, err
}
