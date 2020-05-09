/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package zbec

import (
    "bytes"
    "context"
    "errors"
    "math/rand"
    "reflect"
    "sync"
    "time"

    "github.com/vmihailenco/msgpack"
    "github.com/zlyuancn/zerrors"
    "github.com/zlyuancn/zlog2"
    "github.com/zlyuancn/zsingleflight"

    "github.com/zlyuancn/zbec/cachedb"
    "github.com/zlyuancn/zbec/cachedb/go_cache"
    "github.com/zlyuancn/zbec/cachedb/nocache"
    "github.com/zlyuancn/zbec/errs"
    "github.com/zlyuancn/zbec/query"
)

var (
    // db加载函数不存在或为空
    ErrLoaderFnNotExists = errs.ErrLoaderFnNotExists
    // 缓存或db加载的条目不存在应该返回这个错误
    ErrNoEntry = errs.ErrNoEntry
    // 由缓存保存的ErrNoEntry错误
    NoEntry = errs.NoEntry
)

const (
    // 默认本地缓存有效时间
    DefaultLocalCacheExpire = time.Second
    // 默认缓存空条目有效时间
    DefaultCacheNoEntryExpire = time.Second * 5
)

type Query = query.Query

var NewQuery = query.NewQuery

type BECache struct {
    cdb cachedb.ICacheDB // 缓存数据库

    local_cdb    cachedb.ICacheDB // 本地缓存
    local_cdb_ex time.Duration    // 本地缓存有效时间

    cache_no_entry    bool          // 是否缓存空条目
    cache_no_entry_ex time.Duration // 缓存空条目有效时间

    default_ex    time.Duration // 默认缓存开始时间
    default_endex time.Duration // 默认缓存结束时间

    sf      ISingleFlight      // 单飞
    loaders map[string]ILoader // 加载器配置
    mx      sync.RWMutex       // 对注册的加载器加锁
    log     ILoger             // 日志组件

    deepcopy_result bool // 对结果进行深拷贝
}

func New(c cachedb.ICacheDB, opts ...Option) *BECache {
    m := &BECache{
        cdb: c,

        local_cdb:    nocache.New(),
        local_cdb_ex: DefaultLocalCacheExpire,

        cache_no_entry:    true,
        cache_no_entry_ex: DefaultCacheNoEntryExpire,

        default_ex:    0,
        default_endex: 0,

        sf:      zsingleflight.New(),
        loaders: make(map[string]ILoader),
        log:     zlog2.DefaultLogger,
    }

    for _, o := range opts {
        o(m)
    }
    return m
}

func NewOfGoCache(cleanupInterval time.Duration, opts ...Option) *BECache {
    return New(go_cache.NewGoCache(cleanupInterval), opts...)
}

func NewOfNoCache(opts ...Option) *BECache {
    return New(nocache.New(), opts...)
}

// 设置, 仅用于初始化设置, 正式使用时不应该再调用这个方法
func (m *BECache) SetOptions(opts ...Option) {
    for _, o := range opts {
        o(m)
    }
}

// 为空间注册加载器, 空间名为加载器名, 已注册的空间会被新的加载器替换掉
func (m *BECache) RegisterLoader(loader ILoader) {
    m.mx.Lock()
    m.loaders[loader.Name()] = loader
    m.mx.Unlock()
}

// 获取加载器
func (m *BECache) getLoader(space string) ILoader {
    m.mx.RLock()
    s := m.loaders[space]
    m.mx.RUnlock()
    return s
}

func (m *BECache) cacheGet(query *Query, a interface{}) (interface{}, error) {
    out, err := m.local_cdb.Get(query, a)
    if err == nil || err == NoEntry {
        return out, err
    }

    out, err = m.cdb.Get(query, a)
    if err == nil {
        _ = m.local_cdb.Set(query, out, m.local_cdb_ex)
        return out, nil
    }
    if err == NoEntry {
        _ = m.local_cdb.Set(query, NoEntry, m.local_cdb_ex)
        return nil, NoEntry
    }
    if err == ErrNoEntry {
        return nil, ErrNoEntry
    }
    return nil, zerrors.WithMessage(err, "缓存加载失败")
}
func (m *BECache) cacheSet(query *Query, a interface{}, loader ILoader) {
    _ = m.local_cdb.Set(query, a, m.local_cdb_ex)

    var ex time.Duration
    if a == NoEntry {
        if !m.cache_no_entry {
            return
        }
        ex = m.cache_no_entry_ex
    } else {
        ex = loader.Expire()
        if ex == -1 {
            ex = makeExpire(m.default_ex, m.default_endex)
        }
    }

    if e := m.cdb.Set(query, a, ex); e != nil {
        m.log.Warn(zerrors.WithMessagef(e, "缓存失败<%s>", query.FullPath()))
    }
}
func (m *BECache) cacheDel(query *Query) error {
    _ = m.local_cdb.Del(query)
    return m.cdb.Del(query)
}
func (m *BECache) cacheDelSpace(space string) error {
    _ = m.local_cdb.DelSpaceData(space)
    return m.cdb.DelSpaceData(space)
}

// 从db加载
func (m *BECache) loadDB(query *Query, loader ILoader, delCacheOnErr bool) (interface{}, error) {
    if loader == nil {
        return nil, zerrors.NewSimplef("<%s>加载器为nil", query.Space())
    }

    a, err := loader.Load(query)

    if err == nil {
        if a == nil {
            return nil, zerrors.New("db加载结果不能为nil")
        }
        m.cacheSet(query, a, loader)
        return a, nil
    }

    if err == ErrNoEntry {
        m.cacheSet(query, NoEntry, loader)
        return nil, ErrNoEntry
    }

    if delCacheOnErr {
        if e := m.cdb.Del(query); e != nil { // 从db加载失败时从缓存删除
            m.log.Warn(zerrors.WithMessagef(e, "db加载失败后删除缓存失败<%s>", query.FullPath()))
        }
    }
    return nil, zerrors.WithMessage(err, "db加载失败")
}

// 获取数据, 无数据时空间未注册加载器会返回错误
func (m *BECache) Get(query *Query, a interface{}) error {
    return m.GetWithContext(nil, query, a)
}

// 获取数据, 无数据时空间未注册加载器会返回错误
func (m *BECache) GetWithContext(ctx context.Context, query *Query, a interface{}) error {
    space := m.getLoader(query.Space())
    return m.GetWithLoader(ctx, query, a, space)
}

// 获取数据, 缓存数据不存在时使用指定加载器获取数据
func (m *BECache) GetWithLoader(ctx context.Context, query *Query, a interface{}, loader ILoader) (err error) {
    return doFnWithContext(ctx, func() error {
        return m.getWithLoader(query, a, loader)
    })
}

// 获取数据, 缓存数据不存在时使用指定加载函数获取数据
func (m *BECache) GetWithLoaderFn(ctx context.Context, query *Query, a interface{}, fn LoaderFn) (err error) {
    return m.GetWithLoader(ctx, query, a, NewLoader(fn))
}

func (m *BECache) getWithLoader(query *Query, a interface{}, loader ILoader) error {
    // 同时只能有一个goroutine在获取数据,其它goroutine直接等待结果
    out, err := m.sf.Do(query.FullPath(), func() (interface{}, error) {
        out, err := m.query(query, a, loader)
        if err != nil {
            return nil, err
        }
        if out == nil {
            return nil, nil
        }

        if m.deepcopy_result {
            var buf bytes.Buffer
            err = msgpack.NewEncoder(&buf).Encode(out)
            return buf.Bytes(), err
        }
        return reflect.Indirect(reflect.ValueOf(out)), err
    })

    if err != nil {
        if err == NoEntry {
            err = ErrNoEntry
        }
        return zerrors.WithMessagef(err, "加载失败<%s>", query.FullPath())
    }

    if out == nil {
        return errors.New("未对nil数据做处理")
    }

    if m.deepcopy_result {
        return msgpack.NewDecoder(bytes.NewReader(out.([]byte))).Decode(a)
    }

    reflect.ValueOf(a).Elem().Set(out.(reflect.Value))
    return nil
}

func (m *BECache) query(query *Query, a interface{}, loader ILoader) (interface{}, error) {
    out, gerr := m.cacheGet(query, a)
    if gerr == nil || gerr == NoEntry {
        return out, gerr
    }

    out, lerr := m.loadDB(query, loader, false)
    if lerr == nil {
        return out, lerr
    }

    if gerr != ErrNoEntry { // 有效的错误
        return nil, zerrors.WithMessage(gerr, lerr.Error())
    }
    return nil, lerr
}

// 删除指定数据
func (m *BECache) DelData(query *Query) error {
    return m.cacheDel(query)
}

// 删除指定数据
func (m *BECache) DelDataWithContext(ctx context.Context, query *Query) (err error) {
    return doFnWithContext(ctx, func() error {
        return m.cacheDel(query)
    })
}

// 删除空间数据
func (m *BECache) DelSpaceData(space string) error {
    return m.cacheDelSpace(space)
}

// 删除空间数据
func (m *BECache) DelSpaceDataWithContext(ctx context.Context, space string) error {
    return doFnWithContext(ctx, func() error {
        return m.cacheDelSpace(space)
    })
}

// 设置数据到缓存
func (m *BECache) Set(query *Query, a interface{}, ex ...time.Duration) error {
    return m.SetWithContext(nil, query, a, ex...)
}

// 设置数据到缓存
func (m *BECache) SetWithContext(ctx context.Context, query *Query, a interface{}, ex ...time.Duration) error {
    return doFnWithContext(ctx, func() error {
        var expire = time.Duration(-1)
        if len(ex) > 0 {
            expire = ex[0]
        }

        if a == NoEntry {
            if !m.cache_no_entry {
                _ = m.local_cdb.Set(query, a, m.local_cdb_ex)
                return nil
            }
            expire = m.cache_no_entry_ex
        } else if expire == -1 {
            expire = makeExpire(m.default_ex, m.default_endex)
        }

        if e := m.cdb.Set(query, a, expire); e != nil {
            return zerrors.WithMessagef(e, "缓存失败<%s>", query.FullPath())
        }
        _ = m.local_cdb.Set(query, a, m.local_cdb_ex)
        return nil
    })
}

// 为一个函数添加ctx
func doFnWithContext(ctx context.Context, fn func() error) (err error) {
    if ctx == nil || ctx == context.Background() || ctx == context.TODO() {
        return fn()
    }

    done := make(chan struct{})
    go func() {
        err = fn()
        done <- struct{}{}
    }()

    select {
    case <-done:
        return err
    case <-ctx.Done():
        return ctx.Err()
    }
}

func makeExpire(ex, endex time.Duration) time.Duration {
    if ex == -1 {
        return -1
    }

    if endex == 0 {
        if ex == 0 {
            return 0
        }
        return ex
    }

    return time.Duration(rand.Int63())%(endex-ex) + (ex)
}
