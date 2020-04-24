/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/15
   Description :
-------------------------------------------------
*/

package redis_hash

import (
    "crypto/md5"
    "encoding/hex"
    "time"

    "github.com/afex/hystrix-go/hystrix"
    rredis "github.com/go-redis/redis"
    "github.com/zlyuancn/zerrors"

    "github.com/zlyuancn/zbec/cachedb"
    "github.com/zlyuancn/zbec/codec"
    "github.com/zlyuancn/zbec/errs"
    "github.com/zlyuancn/zbec/query"
)

var _ cachedb.ICacheDB = (*redisWrap)(nil)

type redisWrap struct {
    cdb        rredis.UniversalClient
    codec      codec.ICodec
    md5_params bool
    qfname     string // qf是断路器符号
}

func Wrap(db rredis.UniversalClient, opts ...Option) cachedb.ICacheDB {
    m := &redisWrap{
        cdb:        db,
        codec:      codec.GetCodec(codec.DefaultCodecType),
        md5_params: true,
    }
    for _, o := range opts {
        o(m)
    }
    return m
}

func (m *redisWrap) Set(query *query.Query, v interface{}, ex time.Duration) error {
    if v == errs.NoEntry {
        return m.do(func() error {
            return m.cdb.HSet(query.Space(), m.makeKey(query), []byte{}).Err()
        })
    }

    bs, err := m.codec.Encode(v)
    if err != nil {
        return zerrors.WrapSimplef(err, "编码失败 %T", v)
    }
    return m.do(func() error {
        return m.cdb.HSet(query.Space(), m.makeKey(query), bs).Err()
    })
}

func (m *redisWrap) Get(query *query.Query, a interface{}) (interface{}, error) {
    var data []byte
    var err error
    empty := false
    err = m.do(func() error {
        bs, e := m.cdb.HGet(query.Space(), m.makeKey(query)).Bytes()
        data = bs
        if e == rredis.Nil {
            empty = true
            return nil
        }
        return e
    })

    if err != nil {
        return nil, zerrors.WithSimple(err)
    }
    if empty {
        return nil, errs.ErrNoEntry
    }
    if len(data) == 0 {
        return nil, errs.NoEntry
    }

    err = m.codec.Decode(data, a)
    if err != nil {
        return nil, zerrors.WrapSimplef(err, "解码失败 %T", a)
    }

    return a, nil
}

func (m *redisWrap) Del(query *query.Query) error {
    return m.do(func() error {
        err := m.cdb.HDel(query.Space(), m.makeKey(query)).Err()
        if err == rredis.Nil {
            return nil
        }
        return err
    })
}

func (m *redisWrap) DelSpaceData(space string) error {
    return m.do(func() error {
        err := m.cdb.Del(space).Err()
        if err == rredis.Nil {
            return nil
        }
        return err
    })
}

func (m *redisWrap) makeKey(query *query.Query) string {
    if m.md5_params {
        return string(makeMd5(query.Path()))
    }
    return query.Path()
}

func makeMd5(text string) []byte {
    m := md5.New()
    m.Write([]byte(text))
    src := m.Sum(nil)
    dst := make([]byte, hex.EncodedLen(len(src)))
    hex.Encode(dst, src)
    return dst
}

func (m *redisWrap) do(fn func() error) error {
    if m.qfname == "" {
        return fn()
    }
    return hystrix.Do(m.qfname, fn, nil)
}
