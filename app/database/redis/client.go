/*
 * @Author: cedric.jia
 * @Date: 2021-04-04 18:03:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-04 22:29:51
 */

package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
)

var clientInit sync.Once
var client *redisClient
var rdb *redis.Client

type redisClient struct {
	rdb *redis.Client
}

type RedisKeyNotExistsError struct {
	key string
}

func (err RedisKeyNotExistsError) Error() string {
	return fmt.Sprintf("redis key does not exist [key: %s]", err.key)
}

func IsRedisKeyNotExistsError(err error) bool {
	return err == redis.Nil
}

func Client() *redisClient {
	clientInit.Do(func() {
		client := &redisClient{}
		rdb = redis.NewClient(&redis.Options{
			Addr:     "localhost:6380",
			Password: "",
			DB:       0,
		})
		client.rdb = rdb
	})
	return client
}

func (c *redisClient) SetString(key, val string) error {
	ctx := context.Background()
	if err := rdb.Set(ctx, key, val, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (c *redisClient) GetString(key string) (string, error) {
	ctx := context.Background()
	val, err := rdb.Get(ctx, key).Result()
	if IsRedisKeyNotExistsError(err) {
		return "", RedisKeyNotExistsError{key: key}
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func (c *redisClient) Delete(key string) bool {
	ctx := context.Background()
	cmd := rdb.Del(ctx, key).Val()
	return cmd == 1
}

func (c *redisClient) Exist(key string) bool {
	ctx := context.Background()
	exist := rdb.Exists(ctx, key).Val()
	return exist == 1
}
