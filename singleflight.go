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

type noSingalFlight struct{}

// 一个关闭并发控制的ISingleFlight
func NoSingalFlight() ISingleFlight {
    return new(noSingalFlight)
}

func (*noSingalFlight) Do(_ string, fn func() (interface{}, error)) (interface{}, error) {
    return fn()
}
