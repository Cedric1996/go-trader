/*
 * @Author: cedric.jia
 * @Date: 2021-08-26 12:24:45
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-29 21:26:46
 */

package service

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
)

const (
	PortfolioPath = ".result/"
)

type Position struct {
	positionType string  `json:"position_type"`
	code         string  `json:"code"`
	volume       float64 `json:"volume"`
	price        float64 `json:"price"`
	profitPrice  float64 `json:"profit_price"`
	lossPrice    float64 `json:"loss_price"`
}

func ReadPortfolio(fileName string) ([]*models.Position, error) {
	return nil, nil
}

func GetPortfolio(fileName string) error {
	fileName = PortfolioPath + fileName
	portfolio, err := models.GetPortfolio(1)
	if err != nil {
		return err
	}
	data, err := json.Marshal(portfolio)
	if err != nil {
		return err
	}
	return writeJSON(fileName, data)
}

func newLongPosition(fileName string) error {
	path := PortfolioPath
	if fileName == "" {
		fileName = "long.json"
	}
	path += fileName
	data := readJSON(path, func(data map[string]interface{}) bool {
		t, ok := data["position_type"]
		if !ok || t != "long" {
			return false
		}
		return true
	})
	if err := models.NewPositions(data); err != nil {
		return err
	}
	// portfolio, err := models.CalLongPosition()
	return nil
}

func readJSON(fileName string, filter func(map[string]interface{}) bool) []map[string]interface{} {
	file, _ := os.Open(fileName)
	defer file.Close()

	decoder := json.NewDecoder(file)

	filteredData := []map[string]interface{}{}

	// Read the array open bracket
	decoder.Token()

	data := map[string]interface{}{}
	for decoder.More() {
		decoder.Decode(&data)

		if filter(data) {
			filteredData = append(filteredData, data)
		}
	}

	return filteredData
}

func writeJSON(fileName string, data []byte) error {
	format := "2006-0102-1504"
	t := time.Now().Format(format)
	fileName = fmt.Sprintf("%s_%s.json", fileName, t)
	file, _ := os.OpenFile(fileName, os.O_CREATE, os.ModePerm)
	defer file.Close()
	data, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.Encode(data)
	return nil
}

func syncPortfolio() error {
	positions, err := models.GetHoldPosition()
	if err != nil {
		return err
	}
	portfolio, err := models.GetPortfolio(1)
	if err != nil || portfolio == nil {
		return err
	}
	for _, position := range positions {
		c := &ctx.Context{}
		if err := fetcher.GetCurrentPrice(c, position.Code); err != nil {
			return err
		}
		price := models.ParseCurrentPrice(c)
		position.Price = price
	}
	return nil
}
