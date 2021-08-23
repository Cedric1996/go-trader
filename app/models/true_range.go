/*
 * @Author: cedric.jia
 * @Date: 2021-08-22 17:12:10
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-22 17:20:13
 */

package models

type TrueRange struct {
	Code      string  `bson:"code, omitempty"`
	Date      string  `bson:"date"`
	Timestamp int64   `bson:"timestamp, omitempty"`
	TR        float64 `bson:"tr, omitempty"`
	ATR       float64 `bson:"atr, omitempty"`
}

func RemoveTr(t int64) error {
	return RemoveMany(t, "atr")
}

func InsertTrueRange(data []interface{}) error {
	return InsertMany(data, "true_range")
}
