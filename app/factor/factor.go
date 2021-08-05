/*
 * @Author: cedric.jia
 * @Date: 2021-08-05 14:10:35
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-05 14:12:02
 */

package factor

import ctx "github.cedric1996.com/go-trader/app/context"

type Factor interface {
	Get(ctx.Context) error
	Run() error
}
