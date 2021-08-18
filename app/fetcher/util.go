/*
 * @Author: cedric.jia
 * @Date: 2021-03-14 13:04:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-08-18 12:48:52
 */

package fetcher

import (
	"fmt"
	"strings"
	"sync"
	"time"

	ctx "github.cedric1996.com/go-trader/app/context"
)

var (
	token     string
	tokenInit sync.Once
)

func Token() string {
	tokenInit.Do(func() {
		c := &ctx.Context{}
		if err := GetCurrentToken(c); err != nil {
			fmt.Printf("ERROR: GetCurrentToken error: %s\n", err)
			return
		}
		token = c.ResBody.GetNoKeyVals()[0]
	})
	return token
}

func PostRefDate() string {
	t := strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	return t
}
