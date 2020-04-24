/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/24
   Description :
-------------------------------------------------
*/

package main

import (
    "errors"
    "fmt"

    "github.com/zlyuancn/zbec"
)

func main() {
    // 创建缓存服务
    bec := zbec.NewOfGoCache(0)

    // 注册加载器
    bec.RegisterLoader(zbec.NewNameLoader("test", func(query *zbec.Query) (interface{}, error) {
        if len(query.Params()) == 0 {
            return nil, errors.New("没有传入参数")
        }
        return "hello" + query.Params()[0], nil
    }))

    var a string
    _ = bec.Get(zbec.NewQuery("test", "world"), &a)

    fmt.Println(a)
}
