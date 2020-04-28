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

    "github.com/zlyuancn/zbec/cachedb/go_cache"
    "github.com/zlyuancn/zbec/cachedb/nocache"
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
func WithLocalCache(local_cache bool, ex ...time.Duration) Option {
    return func(m *BECache) {
        if local_cache {
            m.local_cdb = go_cache.NewGoCache(0)
        } else {
            m.local_cdb = nocache.New()
        }

        m.local_cdb_ex = DefaultLocalCacheExpire
        if len(ex) > 0 && ex[0] > 0 {
            m.local_cdb_ex = ex[0]
        }
    }
}

// 设置缓存空条目
func WithCacheNoEntry(cache_no_entry bool, ex ...time.Duration) Option {
    return func(m *BECache) {
        m.cache_no_entry = cache_no_entry
        m.cache_no_entry_ex = DefaultCacheNoEntryExpire
        if len(ex) > 0 && ex[0] > 0 {
            m.cache_no_entry_ex = ex[0]
        }
    }
}

// 对结果进行深拷贝, 每次获取同一个key的结果将分别占有一个内存空间, 代价是性能有所降低
func WithDeepcopyResult(b bool) Option {
    return func(m *BECache) {
        m.deepcopy_result = b
    }
}

// 设置默认缓存时间
func WithDefaultExpire(ex time.Duration, endex ...time.Duration) Option {
    return func(m *BECache) {
        if ex < 1 {
            ex = 0
        }

        m.default_ex = ex
        if len(endex) > 0 {
            m.default_endex = endex[0]
        }
    }
}
