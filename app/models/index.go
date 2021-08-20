/*
 * @Author: cedric.jia
 * @Date: 2021-08-18 19:18:28
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-20 15:07:05
 */

package models

type HighLowIndex struct {
	Date      string `bson:"date,omitempty"`
	Timestamp int64  `bson:"timestamp, omitempty"`
	High      int64  `bson:"high,omitempty"`
	Low       int64  `bson:"low,omitempty"`
	Index     int64  `bson:"index,omitempty"`
}

func InsertHighLowIndex(datas []interface{}) error {
	return InsertMany(datas, "high_low_index")
}

func RemoveHighLowIndex(t int64) error {
	return RemoveMany(t, "high_low_index")
}
