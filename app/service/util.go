/*
 * @Author: cedric.jia
 * @Date: 2021-07-26 20:33:47
 * @Last Modified by: cedric.jia
 * @Last Modified time: 2021-07-26 20:35:07
 */

package service

import (
	"strings"
	"time"
)

func today() string {
	t:= strings.Split(time.Now().Format(time.RFC3339), "T")[0]
	return t
}
