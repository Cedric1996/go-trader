/*
 * @Author: cedric.jia
 * @Date: 2021-04-04 18:18:43
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-04-04 18:57:14
 */

package database

type DB interface {
	Set() error
	Get() error
	Delete() error
}
