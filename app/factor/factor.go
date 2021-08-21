/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:35
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-20 18:10:23
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

func CleanFactorByDate(date string) error {
	t := util.ToTimeStamp(date)
	dates, err := models.GetTradeDay(true, 1, t)
	if err != nil {
		return err
	}
	if len(dates) != 1 || dates[0].Timestamp != t {
		return fmt.Errorf("date: %s has no date to clean", date)
	}
	if err := models.RemoveTradeDay(t); err != nil {
		return err
	}
	if err := models.DeleteStockPriceDayByDay(t); err != nil {
		return err
	}
	factors := []Factor{
		NewEmaFactor(date, 1),
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

func InitFactorByDate(date string) error {
	factors := []Factor{
		NewEmaFactor(date, 1),
		NewHighestFactor("highest", date, 120, true),
		NewHighestFactor("lowest", date, 120, false),
		// NewHighLowIndexFactor("nh_nl", date, false),
		NewRpsFactor("rps", 120, 85, date),
		NewTrendFactor(date, 60, 0.95, 0.75, 2.0, 80),
	}
	for _, f := range factors {
		if err := f.Run(); err != nil {
			return err
		}
	}
	return nil
}
