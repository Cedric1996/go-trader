/*
 * @Author: cedric.jia
 * @Date: 2021-04-24 17:54:29
 * @Last Modified by:   cedric.jia
 * @Last Modified time: 2021-04-24 17:54:29
 */

package service

import (
	"fmt"

	ctx "github.cedric1996.com/go-trader/app/context"
	"github.cedric1996.com/go-trader/app/fetcher"
	"github.cedric1996.com/go-trader/app/models"
)

// Count should not be greater than 5000.
func GetStockInfo(code string) error {
	c := &ctx.Context{}
	if err := fetcher.GetSecurityInfo(c, code); err != nil {
		fmt.Printf("ERROR: GetSecurityInfo error: %s\n", err)
		return err
	}
	if err := models.InsertStockInfo(c); err != nil {
		return err
	}
	return nil
}
