/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/12
   Description :
-------------------------------------------------
*/

package go_cache

import (
    "sync"
    "time"

    "github.com/patrickmn/go-cache"

    "github.com/zlyuancn/zbec/cachedb"
    "github.com/zlyuancn/zbec/errs"
    "github.com/zlyuancn/zbec/query"
)

const DefaultCleanupInterval = time.Minute * 5

var _ cachedb.ICacheDB = (*goCache)(nil)

type goCache struct {
    cdbs map[string]*cache.Cache
    mx   sync.RWMutex

    // 每隔一段时间后清理过期的key
    cleanupInterval time.Duration
}

func NewGoCache(cleanupInterval time.Duration) cachedb.ICacheDB {
    if cleanupInterval <= 0 {
        cleanupInterval = DefaultCleanupInterval
    }
    a := &goCache{
        cdbs:            make(map[string]*cache.Cache),
        cleanupInterval: cleanupInterval,
    }
    return a
}

func (m *goCache) getCache(name string) *cache.Cache {
    m.mx.RLock()
    c, ok := m.cdbs[name]
    m.mx.RUnlock()

    if ok {
        return c
    }

    m.mx.Lock()
    if c, ok = m.cdbs[name]; ok {
        m.mx.Unlock()
        return c
    }

    c = cache.New(0, m.cleanupInterval)
    m.cdbs[name] = c
    m.mx.Unlock()
    return c
}

func (m *goCache) Set(query *query.Query, v interface{}, ex time.Duration) error {
    c := m.getCache(query.Space())
    c.Set(query.Path(), v, ex)
    return nil
}

func (m *goCache) Get(query *query.Query, a interface{}) (interface{}, error) {
    m.mx.RLock()
    c, ok := m.cdbs[query.Space()]
    m.mx.RUnlock()

    if !ok {
        return nil, errs.ErrNoEntry
    }

    out, ok := c.Get(query.Path())
    if !ok {
        return nil, errs.ErrNoEntry
    }

    if out == errs.NoEntry {
        return nil, errs.NoEntry
    }
    return out, nil
}

func (m *goCache) Del(query *query.Query) error {
    m.mx.RLock()
    c, ok := m.cdbs[query.Space()]
    m.mx.RUnlock()

    if !ok {
        return nil
    }

    c.Delete(query.Path())
    return nil
}

func (m *goCache) DelSpaceData(space string) error {
    m.mx.Lock()
    delete(m.cdbs, space)
    m.mx.Unlock()
    return nil
}
