/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package codec

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"

    "github.com/apache/thrift/lib/go/thrift"
    "github.com/gogo/protobuf/proto"
    pb "github.com/golang/protobuf/proto"
    jsoniter "github.com/json-iterator/go"
    "github.com/vmihailenco/msgpack"
)

// 编解码器类型
type CodecType byte

// 默认的编解码器类型
const DefaultCodecType = MsgPack

const (
    // 不进行编解码, 编码解码都直接返回原始数据, 原始数据必须为[]byte或*[]byte
    Byte CodecType = iota
    // 使用go内置的json包进行编解码
    JSON
    // 使用第三方包json-iterator进行编解码
    JsonIterator
    // MsgPack
    MsgPack
    // ProtoBuffer
    ProtoBuffer
    // Thrift
    Thrift
)

// 编解码器
type ICodec interface {
    // 编码
    Encode(i interface{}) ([]byte, error)
    // 解码
    Decode(data []byte, i interface{}) error
}

// 已注册的编解码器
var Codecs = map[CodecType]ICodec{
    Byte:         new(ByteCodec),
    JSON:         new(JSONCodec),
    JsonIterator: new(JSONIteratorCodec),
    ProtoBuffer:  new(PBCodec),
    MsgPack:      new(MsgpackCodec),
    Thrift:       new(ThriftCodec),
}

// 注册自定义编解码器
func RegisterCodec(t CodecType, c ICodec) {
    Codecs[t] = c
}

// 获取编解码器, 如果是未注册的编解码器类型会panic
func GetCodec(t CodecType) ICodec {
    if c, ok := Codecs[t]; ok {
        return c
    }
    panic(fmt.Errorf("未注册的编解码器类型 %v", t))
}

// 不进行编解码
type ByteCodec struct{}

func (ByteCodec) Encode(a interface{}) ([]byte, error) {
    if data, ok := a.([]byte); ok {
        return data, nil
    }
    if data, ok := a.(*[]byte); ok {
        return *data, nil
    }

    return nil, fmt.Errorf("%T 不能转换为 []byte", a)
}

func (ByteCodec) Decode(data []byte, a interface{}) error {
    if ptr_a, ok := a.(*[]byte); ok {
        *ptr_a = data
        return nil
    }
    return fmt.Errorf("%T 不能转换为 *[]byte", a)
}

// 使用go内置的json包进行编解码
type JSONCodec struct{}

func (JSONCodec) Encode(a interface{}) ([]byte, error) {
    return json.Marshal(a)
}

func (JSONCodec) Decode(data []byte, a interface{}) error {
    return json.Unmarshal(data, a)
}

// 使用第三方包json-iterator进行编解码
type JSONIteratorCodec struct{}

func (JSONIteratorCodec) Encode(a interface{}) ([]byte, error) {
    return jsoniter.Marshal(a)
}

func (JSONIteratorCodec) Decode(data []byte, a interface{}) error {
    return jsoniter.Unmarshal(data, a)
}

// Msgpack编解码器
type MsgpackCodec struct{}

func (MsgpackCodec) Encode(a interface{}) ([]byte, error) {
    var buf bytes.Buffer
    enc := msgpack.NewEncoder(&buf)
    enc.UseJSONTag(true) // 如果没有 msgpack 标记, 使用 json 标记
    err := enc.Encode(a)
    return buf.Bytes(), err
}

func (MsgpackCodec) Decode(data []byte, a interface{}) error {
    dec := msgpack.NewDecoder(bytes.NewReader(data))
    dec.UseJSONTag(true) // 如果没有 msgpack 标记, 使用 json 标记
    err := dec.Decode(a)
    return err
}

// ProtoBuffer编解码器
type PBCodec struct{}

func (PBCodec) Encode(a interface{}) ([]byte, error) {
    if m, ok := a.(proto.Marshaler); ok {
        return m.Marshal()
    }

    if m, ok := a.(pb.Message); ok {
        return pb.Marshal(m)
    }

    return nil, fmt.Errorf("%T 不能转换为 proto.Marshaler", a)
}

func (PBCodec) Decode(data []byte, a interface{}) error {
    if m, ok := a.(proto.Unmarshaler); ok {
        return m.Unmarshal(data)
    }

    if m, ok := a.(pb.Message); ok {
        return pb.Unmarshal(data, m)
    }

    return fmt.Errorf("%T 不能转换为 proto.Unmarshaler", a)
}

// Thrift编解码器
type ThriftCodec struct{}

func (c ThriftCodec) Encode(a interface{}) ([]byte, error) {
    b := thrift.NewTMemoryBufferLen(1024)
    p := thrift.NewTBinaryProtocolFactoryDefault().GetProtocol(b)
    t := &thrift.TSerializer{
        Transport: b,
        Protocol:  p,
    }
    _ = t.Transport.Close()
    return t.Write(context.Background(), a.(thrift.TStruct))
}

func (c ThriftCodec) Decode(data []byte, a interface{}) error {
    t := thrift.NewTMemoryBufferLen(1024)
    p := thrift.NewTBinaryProtocolFactoryDefault().GetProtocol(t)
    d := &thrift.TDeserializer{
        Transport: t,
        Protocol:  p,
    }
    _ = d.Transport.Close()
    return d.Read(a.(thrift.TStruct), data)
}
