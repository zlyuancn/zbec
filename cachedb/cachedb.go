/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/12
   Description :
-------------------------------------------------
*/

package cachedb

import (
    "time"

    "github.com/zlyuancn/zbec/query"
)

// 缓存数据库接口
type ICacheDB interface {
    // 设置一个值, ex 为 0 时不应该有过期时间
    // 实现缓存数据库接口的结构应该主动考虑 v 值为 nil(空数据) 如何保存才能在获取时判断它是 NilData
    Set(query *query.Query, v interface{}, ex time.Duration) error
    // 获取一个值, 缓存数据库中不存在应该返回 ErrNoEntry
    // 如果是nil或空数据应该返回 NilData 错误
    Get(query *query.Query, a interface{}) (interface{}, error)
    // 删除一个key
    Del(query *query.Query) error
    // 删除空间数据
    DelSpaceData(space string) error
}
