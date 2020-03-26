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

    rredis "github.com/go-redis/redis"
    "github.com/zlyuancn/zerrors"

    "github.com/zlyuancn/zbec"
    "github.com/zlyuancn/zbec/codec"
)

var _ zbec.ICacheDB = (*RedisWrap)(nil)

type RedisWrap struct {
    cdb rredis.UniversalClient
    c   codec.ICodec
}

func Wrap(db rredis.UniversalClient) *RedisWrap {
    return &RedisWrap{cdb: db, c: codec.GetCodec(codec.DefaultCodecType)}
}

// 设置编解码器
func (m *RedisWrap) SetCodecType(ctype codec.CodecType) *RedisWrap {
    m.c = codec.GetCodec(ctype)
    return m
}

func (m *RedisWrap) Set(query *zbec.Query, v interface{}, ex time.Duration) error {
    if v == nil {
        return m.cdb.Set(makeKey(query), []byte{}, ex).Err()
    }

    bs, err := m.c.Encode(v)
    if err != nil {
        return zerrors.WrapSimplef(err, "编码失败 %T", v)
    }
    return m.cdb.Set(makeKey(query), bs, ex).Err()
}

func (m *RedisWrap) Get(query *zbec.Query, a interface{}) (interface{}, error) {
    bs, err := m.cdb.Get(makeKey(query)).Bytes()
    if err == rredis.Nil {
        return nil, zbec.ErrNoEntry
    }

    if len(bs) == 0 {
        return nil, zbec.NilData
    }

    err = m.c.Decode(bs, a)
    if err != nil {
        return nil, zerrors.WrapSimplef(err, "解码失败 %T", a)
    }

    return a, nil
}

func (m *RedisWrap) Del(query *zbec.Query) error {
    err := m.cdb.Del(makeKey(query)).Err()
    if err == rredis.Nil {
        return nil
    }
    return err
}

func (m *RedisWrap) DelSpaceData(space string) error {
    return zerrors.NewSimple("不提供删除空间数据功能")
}

func makeKey(query *zbec.Query) string {
    var bs bytes.Buffer
    bs.WriteString(query.Space)
    bs.WriteByte(':')
    bs.Write(makeMd5(query.Path()))
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
