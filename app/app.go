/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:02:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-04 18:19:51
 */

package app

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/service"
)

func Init() {
	// mongo.Get()
	// token := fetcher.Token()
	// fmt.Println(token)
	count := fetcher.GetQueryCount()
	fmt.Println(count)
	// stock := fetcher.GetAllSecurities(fetcher.STOCK, "2020-03-12")
	// fmt.Println(stock)
	// stockInfo := fetcher.GetSecurityInfo("600139.XSHG")
	// fmt.Println(stockInfo)
	// bar := fetcher.GetPrice("600139.XSHG", fetcher.Day, 5000)
	// fmt.Println(bar)
	// stocks := fetcher.GetIndexStocks("000300.XSHG","2021-04-02")
	// fmt.Println(stocks)
	// weights := fetcher.GetIndexWeights("000001.XSHE,000002.XSHE", "2021-04-02")
	// fmt.Println(weights)
	data := service.GetFundamentalsData(fetcher.Balance, "000001.XSHE", "2021-04-02")
	fmt.Println(data)
}
