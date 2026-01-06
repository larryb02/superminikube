package store

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func (r *RedisStore) Set(key string, value any) error {
	err := r.client.Set(r.ctx, key, value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisStore) Get(key string) ([]byte, error) {
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

func (r *RedisStore) Delete(key string) error {
	ctx := context.Background()
	_, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisStore) Append(key string, value any) error {
	return nil
}

func NewRedisStore() (*RedisStore, error) {
	ctx := context.Background() // derive this context from parent?
	client := redis.NewClient(&redis.Options{
		Addr: "host.docker.internal:6379", // TODO: make configurable, got so many options to worry about now
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}
