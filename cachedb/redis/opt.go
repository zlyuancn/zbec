/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/8
   Description :
-------------------------------------------------
*/

package redis

import (
    "github.com/zlyuancn/zbec/codec"
)

type Option func(m *RedisWrap)

// 设置编码器
func WithCodecType(ctype codec.CodecType) Option {
    return func(m *RedisWrap) {
        m.codec = codec.GetCodec(ctype)
    }
}

// 设置断路器名, 空名称表示不使用断路器
func WithHystrixName(qfname string) Option {
    return func(m *RedisWrap) {
        m.qfname = qfname
    }
}

// 将query的params做md5, 默认为true
func WithMd5QueryParams(b bool) Option {
    return func(m *RedisWrap) {
        m.md5_params = b
    }
}
