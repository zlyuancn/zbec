/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package redis

import (
    "bytes"
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

var _ cachedb.ICacheDB = (*RedisWrap)(nil)

type RedisWrap struct {
    cdb        rredis.UniversalClient
    codec      codec.ICodec
    md5_params bool
    qfname     string // qf是断路器符号
}

func Wrap(db rredis.UniversalClient, opts ...Option) *RedisWrap {
    m := &RedisWrap{
        cdb:        db,
        codec:      codec.GetCodec(codec.DefaultCodecType),
        md5_params: true,
    }
    for _, o := range opts {
        o(m)
    }
    return m
}

func (m *RedisWrap) Set(query *query.Query, v interface{}, ex time.Duration) error {
    if v == nil {
        return m.do(func() error {
            return m.cdb.Set(m.makeKey(query), []byte{}, ex).Err()
        })
    }

    bs, err := m.codec.Encode(v)
    if err != nil {
        return zerrors.WrapSimplef(err, "编码失败 %T", v)
    }
    return m.do(func() error {
        return m.cdb.Set(m.makeKey(query), bs, ex).Err()
    })
}

func (m *RedisWrap) Get(query *query.Query, a interface{}) (interface{}, error) {
    var data []byte
    var err error
    empty := false
    err = m.do(func() error {
        bs, e := m.cdb.Get(m.makeKey(query)).Bytes()
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

func (m *RedisWrap) Del(query *query.Query) error {
    return m.do(func() error {
        err := m.cdb.Del(m.makeKey(query)).Err()
        if err == rredis.Nil {
            return nil
        }
        return err
    })
}

func (m *RedisWrap) DelSpaceData(space string) error {
    return zerrors.NewSimple("不提供删除空间数据功能")
}

func (m *RedisWrap) makeKey(query *query.Query) string {
    var bs bytes.Buffer
    bs.WriteString(query.Space())
    bs.WriteByte(':')
    if m.md5_params {
        bs.Write(makeMd5(query.Path()))
    } else {
        bs.WriteString(query.Path())
    }
    return bs.String()
}

func makeMd5(text string) []byte {
    m := md5.New()
    m.Write([]byte(text))
    src := m.Sum(nil)
    dst := make([]byte, hex.EncodedLen(len(src)))
    hex.Encode(dst, src)
    return dst
}

func (m *RedisWrap) do(fn func() error) error {
    if m.qfname == "" {
        return fn()
    }
    return hystrix.Do(m.qfname, fn, nil)
}
