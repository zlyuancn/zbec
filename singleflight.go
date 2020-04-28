/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/28
   Description :
-------------------------------------------------
*/

package zbec

type ISingleFlight interface {
    Do(key string, fn func() (interface{}, error)) (interface{}, error)
}
