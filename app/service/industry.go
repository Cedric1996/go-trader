/*
 * @Author: cedric.jia
 * @Date: 2021-07-25 16:35:21
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-09-26 17:32:26
 */
package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
	"github.cedric1996.com/go-trader/app/util"
)

/**
 * industry type:
 * sw_l1: 申万一级行业
 * jq_l1: 聚宽一级行业
 * jq_l2: 聚宽二级行业
 * zjw: 证监会行业
 */
func GetModuleList(moduleType, code string) ([]models.Module, error) {
	c := &ctx.Context{}
	if moduleType == "industry" {
		if err := fetcher.GetIndustryList(c, code); err != nil {
			fmt.Printf("ERROR: GetIndustryList error: %s\n", err)
			return nil, err
		}
	} else {
		if err := fetcher.GetConcepts(c); err != nil {
			fmt.Printf("ERROR: GetModules error: %s\n", err)
			return nil, err
		}
	}
	if err := parseModuleInfo(c); err != nil {
		return nil, err
	}
	return c.Params["modules"].([]models.Module), nil
}

func parseModuleInfo(c *ctx.Context) error {
	resBody := c.ResBody
	code := c.Params["code"]
	res := make([]models.Module, 0)
	if code == "" {
		return fmt.Errorf("parse stock info with error")
	}
	vals := resBody.GetVals()
	for _, val := range vals {
		mod := models.Module{
			Code:      val[0],
			Name:      val[1],
			StartDate: val[2],
		}
		res = append(res, mod)
	}
	c.Params["modules"] = res
	return nil
}

func GetModulesDetail(mod models.Module) ([]interface{}, error) {
	c := &ctx.Context{}
	if err := fetcher.GetConceptStock(c, mod.Code, util.Today()); err != nil {
		return nil, fmt.Errorf("error get concept detail: %s\n", err)
	}
	res := []interface{}{}
	for _, data := range c.ResBody.GetNoKeyVals(){
		res = append(res, models.StockModule{
			Code: data,
			ModuleName: mod.Name,
			StartDate: mod.StartDate,
			Timestamp: util.ParseDate(mod.StartDate).Unix(),
		})
	}
	return res, nil
}