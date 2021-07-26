/*
 * @Author: cedric.jia
 * @Date: 2021-07-26 14:42:38
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-26 15:34:35
 */

package models

type Module struct {
	Code string `bson:"code"`
	Name string `bson:"name"`
	StartDate string `bson:"start_date"`
}

type IndustryModule struct {
	Module  `bson:",inline"`
}

type ConceptModule struct {
	Module  `bson:",inline"`
	Date string `bson:"date"`
}