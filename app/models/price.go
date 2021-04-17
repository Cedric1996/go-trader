/*
 * @Author: cedric.jia
 * @Date: 2021-04-17 17:25:36
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-17 17:37:48
 */

package models

import (
	"github.cedric1996.com/go-trader/app/fetcher"
)

// Price represents basic stock price info.
type Price struct {
	open       float64
	close      float64
	high       float64
	low        float64
	volume     int64
	money      int64
	paused     int64
	high_limit float64
	low_limit  float64
	avg        float64
}

func UpdatePricesByDay(code string, body *fetcher.ResponseBody) error {
	// fmt.Println(body.)
	return nil
}
