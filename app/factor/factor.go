/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:35
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-20 15:43:06
 */

package factor

import (
	"fmt"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/util"
)

type Factor interface {
	Run() error
	Clean() error
}

func CleanDateByDate(date string) error {
	t := util.ToTimeStamp(date)
	dates, err := models.GetTradeDay(true, 1, t)
	if err != nil {
		return err
	}
	if len(dates) != 1 {
		return fmt.Errorf("date: %s has no date to clean", date)
	}
	if err := models.RemoveTradeDay(t); err != nil {
		return err
	}
	if err := models.DeleteStockPriceDayByDay(t); err != nil {
		return err
	}
	factors := []Factor{
		NewEmaFactor("date", 1),
		NewHighLowIndexFactor("nh_nl", date, false),
		NewHighestFactor("highest", date, 120, true),
		NewHighestFactor("lowest", date, 120, false),
		NewRpsFactor("rps", 120, 0, date),
		NewTrendFactor(date, 0, 0, 0, 0, 0),
	}
	for _, f := range factors {
		if err := f.Clean(); err != nil {
			return err
		}
	}
	return nil
}
