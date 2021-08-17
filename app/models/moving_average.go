/*
 * @Author: cedric.jia
 * @Date: 2021-08-17 15:51:51
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-17 16:36:47
 */

package models

type MovingAverage struct {
	Code      string  `bson:"code, omitempty"`
	Date      string  `bson:"date"`
	Timestamp int64   `bson:"timestamp, omitempty"`
	MA_5      float64 `bson:"ma_5, omitempty"`
	MA_10     float64 `bson:"ma_10, omitempty"`
	MA_20     float64 `bson:"ma_20, omitempty"`
	MA_30     float64 `bson:"ma_30, omitempty"`
}

func InsertMovingAverage(datas []interface{}) error {
	return InsertMany(datas, "moving_average")
}
