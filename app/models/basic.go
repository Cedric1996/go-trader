/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:55:00
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 16:33:46
 */

package models

import (
	"context"

	"github.cedric1996.com/go-trader/app/database"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertMany(datas []interface{}, name string) error {
	opts := options.InsertMany()
	_, err := database.Collection(name).InsertMany(context.TODO(), datas, opts)
	if err != nil {
		return err
	}
	return nil
}
