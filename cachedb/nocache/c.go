/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/24
   Description :
-------------------------------------------------
*/

package nocache

import (
    "time"

    "github.com/zlyuancn/zbec/cachedb"
    "github.com/zlyuancn/zbec/errs"
    "github.com/zlyuancn/zbec/query"
)

var _ cachedb.ICacheDB = (*noCache)(nil)

type noCache struct{}

func New() cachedb.ICacheDB { return new(noCache) }

func (noCache) Set(*query.Query, interface{}, time.Duration) error { return nil }

func (noCache) Get(*query.Query, interface{}) (interface{}, error) { return nil, errs.ErrNoEntry }

func (noCache) Del(*query.Query) error { return nil }

func (noCache) DelSpaceData(string) error { return nil }
