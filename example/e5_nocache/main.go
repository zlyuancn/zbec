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
    // 或许你需要测试在没有缓存的情况下服务的状况, 但是不想去拆除大多数地方已经使用了的缓存代码
    // nocache不会缓存任何数据, 我们会始终从加载器或加载函数中读取数据
    bec := zbec.NewOfNoCache()

    var a string
    _ = bec.GetWithLoaderFn(nil, zbec.NewQuery("test"), &a, func(query *zbec.Query) (interface{}, error) {
        return "hello", nil
    })

    fmt.Println(a)
}
