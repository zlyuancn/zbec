/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/12
   Description :  加载器
-------------------------------------------------
*/

package zbec

import (
    "math/rand"
    "time"
)

// 加载器
type ILoader interface {
    // 加载器名
    Name() string
    // 缓存数据不存在时会调用此方法获取数据, 获取的数据会自动缓存
    // 多个进程相同的加载请求同一时刻只有一个进程会调用这个方法, 其他进程会等待结果
    Load(query *Query) (interface{}, error)
    // 数据缓存时会调用这个方法获取缓存时间
    Expire() (ex time.Duration)
}

// db加载函数, 如果是不存在的条目, 应该返回 zbec.ErrNoEntry
type LoaderFn func(query *Query) (interface{}, error)

var _ ILoader = (*Loader)(nil)

// 加载配置
type Loader struct {
    name      string        // 加载器名
    loader    LoaderFn      // 从db加载函数
    ex, endex time.Duration // 有效时间
}

// 创建一个加载器
func NewLoader(loader LoaderFn) *Loader {
    return &Loader{loader: loader}
}

// 创建一个加载器并指定名称
func NewNameLoader(name string, loader LoaderFn) *Loader {
    return &Loader{name: name, loader: loader}
}

func (m *Loader) Name() string {
    return m.name
}

func (m *Loader) Load(query *Query) (interface{}, error) {
    if m.loader == nil {
        return nil, ErrLoaderFnNotExists
    }
    return m.loader(query)
}

func (m *Loader) Expire() time.Duration {
    if m.endex == 0 {
        if m.ex == 0 {
            return 0
        }
        return m.ex
    }

    return time.Duration(rand.Int63())%(m.endex-m.ex) + (m.ex)
}

// 设置加载器名称
func (m *Loader) SetName(name string) *Loader {
    m.name = name
    return m
}

// 设置db加载函数
func (m *Loader) SetLoader(fn LoaderFn) *Loader {
    m.loader = fn
    return m
}

// 设置过期时间
// 如果 ex, endex 都为0, 则永不过期
// 如果 endex 为0, 则过期时间为 ex
// 如果都不为 0, 则过期时间在 [ex, endex] 区间随机
func (m *Loader) SetExpirat(ex, endex time.Duration) *Loader {
    m.ex, m.endex = ex, endex
    return m
}
