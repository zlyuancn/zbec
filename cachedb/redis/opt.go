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
