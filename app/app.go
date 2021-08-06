/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:02:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-05 12:12:40
 */

package app

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/database"
	"github.cedric1996.com/go-trader/app/service"
	"github.com/spf13/viper"
)

func Init() {
	if err := initConfig(); err != nil {
		panic(err)
	}
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	// Handle errors reading the config file
	if err != nil {
		return fmt.Errorf("fatal error config file: %s \n", err)
	}
	database.Init()
	service.Init()
	return nil
}
