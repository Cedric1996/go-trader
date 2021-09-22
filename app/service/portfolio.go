/*
 * @Author: cedric.jia
 * @Date: 2021-08-26 12:24:45
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-22 17:12:43
 */

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/strategy"
	"github.cedric1996.com/go-trader/app/util"
)

const (
	PortfolioPath = ".result/"
)

type position struct {
	Code         string  `json:"code"`
	Price        float64 `json:"price"`
	BeginAt 	 string `json:"begin"`
}

type portfolio struct {
	Data []*position `json:"data"`
}

func ReadPortfolio(fileName string) ([]*models.Position, error) {
	return nil, nil
}

func GetPortfolio(fileName string, t int64) error {
	if err := syncPortfolio(); err != nil {
		return err
	}
	path :=  PortfolioPath + fileName
	file, _ := ioutil.ReadFile(path)
	var po portfolio
	if err := json.Unmarshal([]byte(file), &po);err != nil {
		return err
	}
	date, _ := models.GetTradeDay(true, 1 ,t)
	if len(date) == 0 {
		return errors.New("invalid trade date")
	}
	fmt.Printf("持仓日期: %s\n", util.ToDate(t))
	w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	fmt.Fprintln(w, "股票代码\t名称\t止盈\t止损\t")
	for _ , pos := range po.Data {
		res, err := strategy.HighestRpsPos(pos.Code,  util.ParseDate(pos.BeginAt).Unix(), t)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\t\n", res.Code, res.Name, res.SellPrice, res.LossPrice)
	}
	w.Flush()
	return nil
}

func GetPositionSignal() error {
	positions, err := models.GetHoldPosition()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 5, 5, 10, ' ', 0)
	fmt.Fprintln(w, "Code\tName\tPrice\tLossPrice\t")
	for _, pos := range positions {
		sec, err := models.GetSecurityByCode(pos.Code)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "%s\t%s\t%.2f\t%.2f\t\n", pos.Code, sec.DisplayName, pos.Price, pos.LossPrice)
	}
	w.Flush()
	return nil
}

func NewPosition(fileName string, isShort bool) error {
	// path := PortfolioPath
	// path += fileName
	// positionType := "open"
	// if isShort {
	// 	positionType = "close"
	// }
	// data := readJSON(path, func(data map[string]interface{}) bool {
	// 	t, ok := data["position_type"]
	// 	if !ok || t != positionType {
	// 		return false
	// 	}
	// 	return true
	// })
	// if data == nil {
	// 	return fmt.Errorf("read position file error: %v", path)
	// }
	// if err := models.OpenPositions(data); err != nil {
	// 	return err
	// }
	// if err := syncPortfolio(); err != nil {
	// 	return err
	// }
	return nil
}

func readJSON(fileName string, filter func(map[string]interface{}) bool) []interface{} {
	file, _ := ioutil.ReadFile(fileName)
	data := map[string]interface{}{}
	_ = json.Unmarshal([]byte(file), &data)
	if filter(data) {
		return data["data"].([]interface{})
	}
	return nil
}

func writeJSON(fileName string, data []byte) error {
	format := "2006-0102-1504"
	t := time.Now().Format(format)
	fileName = fmt.Sprintf("%s_%s.json", fileName, t)
	if err := ioutil.WriteFile(fileName, data, 0666); err != nil {
		return err
	}
	return nil
}

func syncPortfolio() error {
	positions, err := models.GetHoldPosition()
	if err != nil {
		return err
	}
	data, err := models.GetPortfolio(1)
	if err != nil || data == nil {
		return err
	}
	portfolio := data[0]
	for _, position := range positions {
		c := &ctx.Context{}
		if err := fetcher.GetPrice(c, position.Code, util.Today(), fetcher.OneMinute, 1); err != nil {
			return err
		}
		price := models.ParseCurrentPrice(c)
		position.Price = price
	}
	portfolio.Positions = positions
	if err := portfolio.CalPortfolio(); err != nil {
		return err
	}
	return nil
}
