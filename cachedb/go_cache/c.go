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

    "github.com/zlyuancn/zbec"
)

const DefaultCleanupInterval = time.Minute * 5

var _ zbec.ICacheDB = (*GoCache)(nil)

type GoCache struct {
    cdbs map[string]*cache.Cache
    mx   sync.RWMutex

    // 每隔一段时间后清理过期的key
    cleanupInterval time.Duration
}

func NewGoCache(cleanupInterval time.Duration) *GoCache {
    if cleanupInterval <= 0 {
        cleanupInterval = DefaultCleanupInterval
    }
    a := &GoCache{
        cdbs:            make(map[string]*cache.Cache),
        cleanupInterval: cleanupInterval,
    }
    return a
}

func (m *GoCache) getCache(name string) *cache.Cache {
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

func (m *GoCache) Set(query *zbec.Query, v interface{}, ex time.Duration) error {
    c := m.getCache(query.Space())
    c.Set(query.Path(), v, ex)
    return nil
}

func (m *GoCache) Get(query *zbec.Query, a interface{}) (interface{}, error) {
    m.mx.RLock()
    c, ok := m.cdbs[query.Space()]
    m.mx.RUnlock()

    if !ok {
        return nil, zbec.ErrNoEntry
    }

    out, ok := c.Get(query.Path())
    if !ok {
        return nil, zbec.ErrNoEntry
    }

    if out == nil {
        return nil, zbec.NilData
    }
    return out, nil
}

func (m *GoCache) Del(query *zbec.Query) error {
    m.mx.RLock()
    c, ok := m.cdbs[query.Space()]
    m.mx.RUnlock()

    if !ok {
        return nil
    }

    c.Delete(query.Path())
    return nil
}

func (m *GoCache) DelSpaceData(space string) error {
    m.mx.Lock()
    delete(m.cdbs, space)
    m.mx.Unlock()
    return nil
}
