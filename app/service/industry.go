/*
 * @Author: cedric.jia
 * @Date: 2021-07-25 16:35:21
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-08 10:04:09
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
func GetModuleList(moduleType, code string) error {
	c := &ctx.Context{}
	if moduleType == "industry" {
		if err := fetcher.GetIndustryList(c, code); err != nil {
			fmt.Printf("ERROR: GetIndustryList error: %s\n", err)
			return err
		}
	} else {
		if err := fetcher.GetConcepts(c); err != nil {
			fmt.Printf("ERROR: GetModules error: %s\n", err)
			return err
		}
	}
	if err := parseModuleInfo(c); err != nil {
		return err
	}
	if err := getModulesDetail(c); err != nil {
		return err
	}
	return nil
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

func getModulesDetail(c *ctx.Context) error {
	modules, has := c.Params["modules"].([]models.Module)
	if !has {
		return fmt.Errorf("error get modules detail")
	}
	modules = modules[0:10]
	for _, module := range modules {
		getModuleDetail(&models.ConceptModule{
			Module: module,
			Date:   util.Today(),
		})
	}
	return nil
}

func getModuleDetail(module *models.ConceptModule) {
	c := &ctx.Context{}
	if err := fetcher.GetConceptStock(c, module.Module.Code, module.Date); err != nil {
		fmt.Printf("error get concept detail: %s\n", err)
		return
	}

	fmt.Printf("get concept detail: %s, %s\n", module, c.ResBody)
}
