package factor

import (
	"errors"
	"fmt"
	"math"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/modules/queue"
	"github.cedric1996.com/go-trader/app/service"
	"github.cedric1996.com/go-trader/app/util"
)

type RpsWeekFactor struct {
	calDate   string `bson:"calDate, omitempty"`
	timestamp int64  `bson:"timestamp, omitempty"`
}

type rpsWeekDatum struct {
	code      string
	calDate   string
	timestamp int64
	rps_250   int64
	rps_120   int64
	rps_60    int64
}

func NewRpsWeekFactor(calDate string) *RpsWeekFactor {
	return &RpsWeekFactor{
		calDate:   calDate,
		timestamp: util.ParseDate(calDate).Unix(),
	}
}

func (f *RpsWeekFactor) Run() error {
	if err := f.execute(); err != nil {
		return err
	}
	return nil
}

// func (f *RpsWeekFactor) Clean() error {
// 	return models.RemoveRpsWeek(f.timestamp)
// }

func (f *RpsWeekFactor) execute() error {
	tmps, err := models.GetRpsDates()
	if err != nil {
		return err
	}
	queue, err := queue.NewQueue("rps_week", f.calDate, 50, 1000, func(data interface{}) (interface{}, error) {
		datum := data.(rpsWeekDatum)
		code := datum.code
		timestamp := datum.timestamp
		highest, err := models.GetHighestList(models.SearchOption{Code: code, BeginAt: timestamp}, "highest_120")
		if err != nil || len(highest) < 120 {
			return nil, errors.New("")
		}
		prices, err := models.GetStockPriceList(models.SearchOption{Code: code, Timestamp: timestamp})
		if err != nil || prices == nil {
			return nil, errors.New("")
		}
		highestPrice := highest[len(highest)-120]
		datas, err := service.InitFundamental(code, f.calDate, 1)
		if err != nil || datas == nil {
			return nil, errors.New("")
		}
		return models.RpsWeek{
			RpsBase: models.RpsBase{
				Code:      code,
				Timestamp: timestamp,
				Date:      datum.calDate,
			},
			Rps_250:   datum.rps_250,
			Rps_120:   datum.rps_120,
			Rps_60:    datum.rps_60,
			Net:       (highestPrice.Price - prices[0].Price.Close) / prices[0].Price.Close,
			MarketCap: datas[0].(models.Valuation).MarketCap,
		}, nil
	}, func(data []interface{}) error {
		if err := models.InsertRpsWeek(data); err != nil {
			return err
		}
		return nil
	})
	for _, tmp := range tmps {
		m, err := service.GetNewRps(tmp.(int64))
		if err != nil {
			continue
		}
		for _, data := range *m {
			queue.Push(rpsWeekDatum{
				code:      data.RpsBase.Code,
				calDate:   data.RpsBase.Date,
				timestamp: data.RpsBase.Timestamp,
				rps_250:   data.Rps_250,
				rps_120:   data.Rps_120,
				rps_60:    data.Rps_60,
			})
		}
	}
	queue.Close()
	return nil
}

func (f *RpsWeekFactor) valuationFilter(code string, marketCap float64) error {
	datas, err := service.InitFundamental(code, f.calDate, 1)
	if err != nil {
		return err
	}
	if len(datas) != 1 {
		return fmt.Errorf("fetch valuation error, code: %v", code)
	}
	val := datas[0].(models.Valuation).MarketCap
	if math.Dim(val, marketCap) < 0.1 {
		return fmt.Errorf("marketCap is less than %v, code: %v", marketCap, code)
	}
	return nil
}
