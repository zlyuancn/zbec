/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/16
   Description :
-------------------------------------------------
*/

package errs

import (
    "errors"
)

var ErrLoaderFnNotExists = errors.New("db加载函数不存在或为空")

// 缓存或db加载的条目不存在应该返回这个错误
var ErrNoEntry = errors.New("条目不存在")

// 由缓存保存的ErrNoEntry错误
var NoEntry = errors.New("空条目")
