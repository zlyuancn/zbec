/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package zbec

import (
    "fmt"
    "strings"
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

    if len(q.Params) == 0 {
        q.path = fmt.Sprintf("%s/%s", q.Method, q.Key)
    } else {
        q.path = fmt.Sprintf("%s/%s?%s", q.Method, q.Key, strings.Join(q.Params, "&"))
    }
    return q.path
}

// 获取包含空间名的完整路径
func (q *Query) FullPath() string {
    if q.full_path != "" {
        return q.full_path
    }

    q.full_path = fmt.Sprintf("%s:%s", q.Space, q.Path())
    return q.full_path
}
