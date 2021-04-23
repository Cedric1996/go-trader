/*
 * @Author: cedric.jia
 * @Date: 2021-04-04 18:03:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-23 23:28:48
 */

package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var clientInit sync.Once
var client *RedisClient
var rdb *redis.Client

type RedisClient struct {
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

func Client() *RedisClient {
	clientInit.Do(func() {
		client := &RedisClient{}
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", viper.GetString("redis.hostname"), viper.GetString("redis.port")),
			Password: "",
			DB:       0,
		})
		client.rdb = rdb
	})
	return client
}

func (c *RedisClient) HSet(key string, member ...string) error {
	ctx := context.Background()
	if err := rdb.HSet(ctx, key, member).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisClient) HExists(key, member string) (bool, error) {
	ctx := context.Background()
	cmd := rdb.HExists(ctx, key, member)
	if cmd.Err() != nil {
		return false, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *RedisClient) HDel(key string, member ...string) error {
	ctx := context.Background()
	if err := rdb.HDel(ctx, key, member...).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisClient) MSet(key string, member interface{}) error {
	ctx := context.Background()
	if err := rdb.MSet(ctx, key, member).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisClient) MGet(key string) (interface{}, error) {
	ctx := context.Background()
	cmd := rdb.MGet(ctx, key)
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}
	return cmd.Val(), nil
}

func (c *RedisClient) Flush(hsetName string) error {
	ctx := context.Background()
	keys, err := c.rdb.HKeys(ctx, hsetName).Result()
	if err != nil {
		return err
	}
	if err = c.rdb.Del(ctx, keys...).Err(); err != nil {
		return err
	}
	return c.rdb.Del(ctx, hsetName).Err()
}
