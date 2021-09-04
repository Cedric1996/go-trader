/*
 * @Author: cedric.jia
 * @Date: 2021-08-30 10:35:17
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-31 22:23:44
 */

package service

import (
	"encoding/json"
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
	// if err := GetNewVcp(tradeDay[0].Timestamp, &codeMap); err != nil {
	// 	return nil, err
	// }
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
