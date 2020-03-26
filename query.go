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
)

// 查询参数
type Query struct {
    path      string
    full_path string
    // 空间名
    Space string
    // 方法
    Method string
    // 主键
    Key string
    // 参数
    Params []string
}

// 创建一个查询结构
func NewQuery(space, method, key string, params ...string) *Query {
    q := &Query{
        Space:  space,
        Method: method,
        Key:    key,
        Params: params,
    }
    q.Check()
    return q
}

// 检查参数
func (q *Query) Check() {
    if q.Space == "" {
        panic("空间名为空")
    }
    if q.Method == "" {
        panic("方法名为空")
    }
}

// 获取路径
func (q *Query) Path() string {
    if q.path != "" {
        return q.path
    }

    var bs bytes.Buffer
    bs.WriteString(q.Method)
    bs.WriteByte(':')
    bs.WriteString(q.Key)
    if len(q.Params) > 0 {
        bs.WriteByte('?')
        for _, v := range q.Params {
            bs.WriteString(v)
            bs.WriteByte('&')
        }
        q.path = string(bs.Bytes()[:bs.Len()-1])
        return q.path
    }
    q.path = bs.String()
    return q.path
}

// 获取包含空间名的完整路径
func (q *Query) FullPath() string {
    if q.full_path != "" {
        return q.full_path
    }

    var bs bytes.Buffer
    bs.WriteString(q.Space)
    bs.WriteByte(':')
    bs.WriteString(q.Path())
    bs.Bytes()
    q.full_path = bs.String()
    return q.full_path
}
