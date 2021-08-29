/*
 * @Author: cedric.jia
 * @Date: 2021-04-07 22:48:55
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-29 21:12:20
 */

package database

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/database/mongodb"
	"github.cedric1996.com/go-trader/app/database/redis"
	"github.com/spf13/viper"
)

var (
	successSet string
	failSet    string
	handleSet  string
)

func Init() {
	successSet = viper.GetString("redis.sets.success")
	failSet = viper.GetString("redis.sets.fail")
	handleSet = viper.GetString("redis.maps.handle")
	mongodb.ConnectMongoClient()
}

func FetchSuccess(member string) error {
	if err := redis.Client().HSet(successSet, member); err != nil {
		return fmt.Errorf("set Fetch Success status error: %v", err)
	}
	return nil
}

func FetchFail(member string) error {
	if err := redis.Client().HSet(failSet, member); err != nil {
		return fmt.Errorf("set Fetch Fail status error: %v", err)
	}
	return nil
}

func IsFetchSuccess(member string) (bool, error) {
	return redis.Client().HExists(successSet, member)
}

func IsFetchFail(member string) (bool, error) {
	return redis.Client().HExists(failSet, member)
}

func HandleFailed(key string) error {
	if err := redis.Client().MSet(handleSet, key); err != nil {
		return fmt.Errorf("set Fetch Success status error: %v", err)
	}
	return nil
}
