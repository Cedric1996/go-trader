/*
 * @Author: cedric.jia
 * @Date: 2021-08-30 10:35:17
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-24 15:45:39
 */

package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/util"
)

func GetVcpByInterval(startDate string, interval int64) (map[string]int, error) {
	t := util.ParseDate(startDate).Unix()
	tradeDay, err := models.GetTradeDay(true, interval, t)
	if err != nil || len(tradeDay) != int(interval) {
		return nil, nil
	}
	vcps, err := models.GetVcp(models.SearchOption{
		BeginAt: tradeDay[interval-1].Timestamp,
		EndAt:   tradeDay[0].Timestamp,
	})
	if err != nil {
		return nil, err
	}
	codeMap := map[string]int{}
	for _, vcp := range vcps {
		_, ok := codeMap[vcp.RpsBase.Code]
		if ok {
			codeMap[vcp.RpsBase.Code] += 1
		} else {
			codeMap[vcp.RpsBase.Code] = 1
		}
	}
	return codeMap, nil
}

func GenerateVcpFile(datas []string) error {
	codes := make([]string, 0)
	for _, v := range datas {
		parts := strings.Split(v, ".")
		prefix := "sh"
		if parts[1] == "XSHE" {
			prefix = "sz"
		}
		codes = append(codes, prefix+parts[0])
	}
	result := make(map[string]interface{})
	result["leek-fund.stocks"] = codes

	data, err := json.Marshal(&result)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(".result/result.json", data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func GetNewVcp(t int64, vcpMap *map[string]int) error {
	newVcps, err := models.GetVcpByDate(t)
	if err != nil {
		return err
	}
	newMap := make(map[string]int)
	mapVal := *vcpMap
	for _, v := range newVcps {
		val := mapVal[v.RpsBase.Code]
		if val != 1 {
			newMap[v.RpsBase.Code] = val
		}
	}
	vcpMap = &newMap
	return nil
}

func GetNewRps(unix int64) (*map[string]*models.Rps, error) {
	tradeDay, err := models.GetTradeDay(true, 6, unix)
	if err != nil || len(tradeDay) != 6 {
		return nil, err
	}
	old, err := models.GetRpsByOpt(models.SearchOption{BeginAt: tradeDay[5].Timestamp, EndAt: tradeDay[1].Timestamp})
	if err != nil || old == nil {
		return nil, errors.New("")
	}
	new, err := models.GetRpsByOpt(models.SearchOption{Timestamp: tradeDay[0].Timestamp})
	if err != nil || new == nil {
		return nil, errors.New("")
	}
	newMap := make(map[string]*models.Rps)
	for _, v := range new {
		if v.Rps_250 >= 90 || v.Rps_120 >= 90 || v.Rps_60 >= 90 {
			newMap[v.RpsBase.Code] = v
		}
	}
	for _, v := range old {
		vaild := v.Rps_250 >= 90 || v.Rps_120 >= 90 || v.Rps_60 >= 90
		if _, ok := newMap[v.RpsBase.Code]; vaild && ok {
			delete(newMap, v.RpsBase.Code)
		}
	}
	return &newMap, nil
}

type highestApproach struct {
	Code  string   `json:"code"`
	Name  string   `json:"name"`
	Dates []string `json:"dates"`
}

func ExportHighestApproach() error {
	securities, err := models.GetAllSecurities()
	if err != nil {
		return err
	}
	results := make([]highestApproach, 0)
	for _, stock := range securities {
		datas, err := models.GetHighestApproach(models.SearchOption{Code: stock.Code})
		if err != nil || datas == nil {
			continue
		}
		dates := []string{}
		for _, data := range datas {
			dates = append(dates, data.RpsBase.Date)
		}
		results = append(results, highestApproach{
			Code:  stock.Code,
			Name:  stock.DisplayName,
			Dates: dates,
		})
	}
	data, err := json.Marshal(&results)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(".result/highest_approach.json", data, os.ModePerm); err != nil {
		return err
	}
	return nil
}
