/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package zbec

import (
    "time"

    "github.com/zlyuancn/zlog2"
)

type Option func(m *BECache)

// 设置日志组件
func WithLogger(log ILoger) Option {
    return func(m *BECache) {
        if log == nil {
            log = zlog2.DefaultLogger
        }
        m.log = log
    }
}

// 设置本地缓存
func WithLocalCache(ex time.Duration, local_cache ICacheDB) Option {
    return func(m *BECache) {
        if ex <= 0 {
            ex = DefaultLocalCacheExpire
        }
        m.local_cdb_ex = ex
        m.local_cdb = local_cache
    }
}

// 设置是否缓存空数据, 注意: 本地缓存一定会缓存空数据
func WithCacheNilData(b bool) Option {
    return func(m *BECache) {
        m.cache_nil = b
    }
}
