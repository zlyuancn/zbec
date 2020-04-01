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
//
// 在缓存和读取值时会将 Space, []Params 用作"综合key"计算
type Query struct {
    // 空间名
    space string
    // 参数
    params []string
    // 附加对象
    a interface{}

    path      string
    full_path string
}

// 创建一个查询结构
func NewQuery(space string, params ...string) *Query {
    if space == "" {
        panic("空间名为空")
    }

    q := &Query{
        space:  space,
        params: params,
    }
    q.makePath()
    q.makeFullPath()
    return q
}

// 获取空间名
func (q *Query) Space() string {
    return q.space
}

// 获取参数
func (q *Query) Params() []string {
    return q.params
}

// 设置附加数据, 这个对象不会参与到"综合key"计算中
//
// 我们在实际开发时, 发现将db加载需要的数据转为query, 然后在db加载函数中将query再转为自己需要的数据比较困难
// 这个功能将解决以上问题, 你可以放心的将你想要告知db加载函数的数据传过去
// 需要注意的是, 同一个"综合key"的附加数据内容必须一致
func (q *Query) SetPayload(a interface{}) *Query {
    q.a = a
    return q
}

// 获取附加数据
func (q *Query) Payload() interface{} {
    return q.a
}

func (q *Query) makePath() {
    if len(q.params) > 0 {
        var bs bytes.Buffer
        bs.WriteByte('?')
        for _, v := range q.params {
            bs.WriteString(v)
            bs.WriteByte('&')
        }
        q.path = string(bs.Bytes()[:bs.Len()-1])
    }
}
func (q *Query) makeFullPath() {
    var bs bytes.Buffer
    bs.WriteString(q.space)
    bs.WriteByte(':')
    bs.WriteString(q.path)
    q.full_path = bs.String()
}

// 获取路径
func (q *Query) Path() string {
    return q.path
}

// 获取包含空间名的完整路径
func (q *Query) FullPath() string {
    return q.full_path
}
