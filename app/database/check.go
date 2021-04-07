/*
 * @Author: cedric.jia
 * @Date: 2021-04-07 22:48:55
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-07 23:05:19
 */

package database

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/database/redis"
	"github.com/spf13/viper"
)

var (
	successSet = viper.GetString("redis.sets.success")
	failSet    = viper.GetString("redis.sets.fail")
)

func FetchSuccess(member string) error {
	if err := redis.Client().SAdd(successSet, member); err != nil {
		return fmt.Errorf("set Fetch Success status error: %v", err)
	}
	return nil
}

func FetchFail(member string) error {
	if err := redis.Client().SAdd(failSet, member); err != nil {
		return fmt.Errorf("set Fetch Fail status error: %v", err)
	}
	return nil
}

func IsFetchSuccess(key, member string) (bool, error) {
	return redis.Client().SIsMember(successSet, member)
}

func IsFetchFail(key, member string) (bool, error) {
	return redis.Client().SIsMember(failSet, member)
}
