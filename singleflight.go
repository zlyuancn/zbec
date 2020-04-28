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

type NoSingalFlight struct{}

func (*NoSingalFlight) Do(_ string, fn func() (interface{}, error)) (interface{}, error) {
    return fn()
}
