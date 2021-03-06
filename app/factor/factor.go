/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:35
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-27 16:56:14
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
	factors := []Factor{
		NewEmaFactor(date, 1),
		NewHighLowIndexFactor("nh_nl", date),
		NewHighestFactor("highest", date, 60),
		NewHighestFactor("highest", date, 120),
		NewRpsFactor("rps", 120, 0, date),
		NewTrendFactor(date, 0, 0, 0, 0),
		NewTrueRangeFactor(date, 13),
		NewHighestRpsFactor(date, 0.95, 2.0),
	}
	return cleanFactor(date, factors)
}

func CleanPosByDate(date string) error {
	factors := []Factor{
		NewHighestFactor("highest", date, 120),
		NewRpsFactor("rps", 120, 0, date),
		NewHighestRpsFactor(date, 0.95, 2.0),
	}
	return cleanFactor(date, factors)
}

func cleanFactor(date string, factors []Factor) error {
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
		NewHighestFactor("highest", date, 60),
		NewHighestFactor("highest", date, 120),
		NewRpsFactor("rps", 120, 85, date),
		// NewHighLowIndexFactor("nh_nl", date),
		// NewTrueRangeFactor(date, 13),
		// NewTrendFactor(date, 60, 0.9, 0.80, 2.0),
		NewHighestRpsFactor(date, 0.95, 2.0),
	}
	return initFactor(factors)
}

func InitPosByDate(date string) error {
	factors := []Factor{
		NewHighestFactor("highest", date, 60),
		NewHighestFactor("highest", date, 120),
		NewTrendFactor(date, 60, 0.9, 0.80, 2.0),
	}
	return initFactor(factors)
}

func initFactor(factors []Factor) error {
	for _, f := range factors {
		if err := f.Run(); err != nil {
			return err
		}
	}
	return nil
}
