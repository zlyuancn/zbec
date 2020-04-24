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

type QueryPayLoad struct {
    A string
}

func main() {
    // 创建缓存服务
    bec := zbec.NewOfGoCache(0)

    // 注册加载器
    bec.RegisterLoader(zbec.NewNameLoader("test", func(query *zbec.Query) (interface{}, error) {
        // 读取查询荷载
        payload := query.Payload().(*QueryPayLoad)
        return "hello" + payload.A, nil
    }))

    // 创建查询荷载
    payload := &QueryPayLoad{A: "world"}

    // 创建一个query, 在缓存不存在时, 我们会根据 spanc 和 params 来定位查询荷载, 所以不同的荷载值需要写入不同的参数
    // 最佳实践: 同一个 space 的查询荷载只能是确定的一个类型, 查询参数根据荷载结构的参数顺序写入
    query := zbec.NewQuery("test", "world").SetPayload(payload)

    var a string
    _ = bec.Get(query, &a)

    fmt.Println(a)
}
