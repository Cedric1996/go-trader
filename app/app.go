/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:02:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-03-14 22:48:51
 */

package app

import (
	"fmt"

	"github.cedric1996.com/eztrader/app/database"
	"github.cedric1996.com/eztrader/app/fetcher"
)

func Init() {
	database.Init()
	// token := fetcher.Token()
	// fmt.Println(token)
	// count := fetcher.GetQueryCount()
	// fmt.Println(count)
	// stock := fetcher.GetAllSecurities(fetcher.STOCK, "2020-03-12")
	// fmt.Println(stock)
	// stockInfo := fetcher.GetSecurityInfo("600139.XSHG")
	// fmt.Println(stockInfo)
	bar := fetcher.GetPrice("600139.XSHG", fetcher.Day, 5000)
	fmt.Println(bar)

}
