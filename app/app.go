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
	// var err error
	// if err := mongo.CreateCollection("test2"); err != nil {
	// 	fmt.Println(err.Error())
	// }
	// if err := s.Insert(); err != nil {
	// 	fmt.Println(err.Error())
	// }
	err := service.GetStockPriceByCode("000001.XSHE")
	// token := fetcher.Token()
	// fmt.Println(token)
	// count := fetcher.GetQueryCount()
	// fmt.Println(count)
	// stock := fetcher.GetAllSecurities(fetcher.STOCK, "2020-03-12")
	// fmt.Println(stock)
	// err = service.GetModuleList("concept", "sw_l3")
	// fmt.Println(err)
	// err = service.GetPricesByDay("600139.XSHG", 1)
	// bar, _ := fetcher.GetPrice("600139.XSHG", fetcher.Day, 1)
	// fmt.Println(err)
	// stock, _ := models.GetStockInfoByCode("600139.XSHG")
	// fmt.Println(stock)
	// stocks := fetcher.GetIndexStocks("000300.XSHG","2021-04-02")
	// fmt.Println(stocks)
	// weights := fetcher.GetIndexWeights("000001.XSHE,000002.XSHE", "2021-04-02")
	// fmt.Println(weights)
	// err = service.GetFundamentalsData(fetcher.Valuation, "000001.XSHE", "")
	// err = models.InitStockTableIndexes()
	fmt.Println(err)
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
