/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/24
   Description :
-------------------------------------------------
*/

package main

import (
    "fmt"

    "github.com/zlyuancn/zbec"
)

func main() {
    // 创建缓存服务
    bec := zbec.NewOfGoCache(0)

    // 注册加载器
    bec.RegisterLoader(zbec.NewNameLoader("test", func(query *zbec.Query) (interface{}, error) {
        return "hello", nil
    }))

    // 创建用户保存结果的变量
    var a string

    // 从缓存加载, 如果加载失败会调用加载器, 随后将结果放入缓存中, 最后将数据写入a
    _ = bec.Get(zbec.NewQuery("test"), &a)

    fmt.Println(a)
}
