/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package zbec

// 日志组件接口
type ILoger interface {
    Info(v ...interface{})
    Warn(v ...interface{})
    Error(v ...interface{})
}
