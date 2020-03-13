/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/3/11
   Description :
-------------------------------------------------
*/

package test

import (
    "bytes"
    "fmt"
    "math/rand"
    "testing"
    "time"

    rredis "github.com/go-redis/redis"
    "github.com/zlyuancn/zerrors"

    "github.com/zlyuancn/zbec"
    "github.com/zlyuancn/zbec/cachedb/go_cache"
    "github.com/zlyuancn/zbec/cachedb/redis"
    "github.com/zlyuancn/zbec/codec"
)

func getRedisClient(on_local_cache bool) *zbec.BECache {
    cdb := rredis.NewClient(&rredis.Options{
        Addr:        "127.0.0.1:6379",
        Password:    "",
        DB:          0,
        PoolSize:    10,
        DialTimeout: time.Second * 3,
    })
    if err := cdb.Ping().Err(); err != nil {
        panic(err)
    }

    bec := zbec.New(redis.Wrap(cdb).SetCodecType(codec.Byte))

    if on_local_cache {
        lcdb := go_cache.NewGoCache(0)
        bec.SetOptions(zbec.WithLocalCache(0, lcdb))
    }
    return bec
}

func getGoCache() *zbec.BECache {
    cdb := go_cache.NewGoCache(0)
    bec := zbec.New(cdb)
    return bec
}

func TestGetAndCache(t *testing.T) {
    space := "test"
    loader := zbec.NewNameLoader(space, func(query *zbec.Query) (i interface{}, err error) {
        s := fmt.Sprintf("%s", query.FullPath())
        return &s, nil
    })

    bec := getGoCache()
    bec.RegisterLoader(loader)

    for i := 0; i < 10; i++ {
        k := fmt.Sprintf("k%d", i)
        v := fmt.Sprintf("v%d", i)
        query := zbec.NewQuery(space, k, v)

        a := new(string)
        err := bec.Get(query, a)
        if err != nil {
            t.Fatalf("%+v", err)
        }

        if *a != query.FullPath() {
            t.Fatalf("收到的值非预期 %s: %s", query.FullPath(), *a)
        }
    }
}

// go test -v -bench "^Benchmark_.+$" -run ^$ -cpu 20,50,100,1000,10000 .
// docker run --rm -v $PWD/../..:/src/app -v /src/gopath:/src/gopath -v /src/gocache:/src/gocache -w /src/app/zbec/test zlyuan/golang:1.13 go test -v -bench "^Benchmark_.+$" -run ^$ -cpu 20,50,100,1000,10000 .

func Benchmark_GoCache1e3(b *testing.B) {
    bec := getGoCache()
    benchmark_any(b, bec, 1e3)
}

func Benchmark_Redis1e3(b *testing.B) {
    bec := getRedisClient(false)
    benchmark_any(b, bec, 1e3)
}

func Benchmark_RedisAndLocalCache1e3(b *testing.B) {
    bec := getRedisClient(true)
    benchmark_any(b, bec, 1e3)
}

func Benchmark_GoCache1e4(b *testing.B) {
    bec := getGoCache()
    benchmark_any(b, bec, 1e4)
}

func Benchmark_Redis1e4(b *testing.B) {
    bec := getRedisClient(false)
    benchmark_any(b, bec, 1e4)
}

func Benchmark_RedisAndLocalCache1e4(b *testing.B) {
    bec := getRedisClient(true)
    benchmark_any(b, bec, 1e4)
}

func Benchmark_GoCache1e5(b *testing.B) {
    bec := getGoCache()
    benchmark_any(b, bec, 1e5)
}

func Benchmark_Redis1e5(b *testing.B) {
    bec := getRedisClient(false)
    benchmark_any(b, bec, 1e5)
}

func Benchmark_RedisAndLocalCache1e5(b *testing.B) {
    bec := getRedisClient(true)
    benchmark_any(b, bec, 1e5)
}

func benchmark_any(b *testing.B, bec *zbec.BECache, max_key_count int) {
    rand.Seed(time.Now().UnixNano())

    byte_len := 512

    space := "benchmark"
    loader := zbec.NewNameLoader(space, func(query *zbec.Query) (interface{}, error) {
        bs := make([]byte, byte_len)
        for i := 0; i < len(bs); i++ {
            bs[i] = byte(rand.Int() % 256)
        }
        return &bs, nil
    })
    bec.RegisterLoader(loader)

    vs := make([][]byte, max_key_count)
    // 超过10000个key不做值检查
    if max_key_count <= 10000 {
        for i := 0; i < max_key_count; i++ {
            k := fmt.Sprintf("%d", i)

            a := new([]byte)
            err := bec.Get(zbec.NewQuery(space, k, ""), a)
            if err != nil {
                b.Fatalf("%+v", zerrors.WrapSimplef(err, "初始化失败: %s", k))
            }

            if len(*a) != byte_len {
                b.Fatalf("%+v", zerrors.NewSimplef("初始化失败, 收到的值长度非预期 %s, %d", k, len(*a)))
            }

            vs[i] = *a
        }
    }

    // 缓存随机key, 因为它本身速度太慢了(117ns/op)
    randv := make([]int, 10000000)
    randvlen := len(randv)
    for i := 0; i < randvlen; i++ {
        randv[i] = rand.Int() % max_key_count
    }

    b.ResetTimer()
    b.RunParallel(func(p *testing.PB) {
        i := 0
        for p.Next() {
            i++

            m := randv[i%randvlen]
            k := fmt.Sprintf("%d", m)

            a := new([]byte)
            err := bec.Get(zbec.NewQuery(space, k, ""), a)
            if err != nil {
                b.Fatalf("%+v", zerrors.WrapSimplef(err, "测试失败: %s", k))
            }

            if len(*a) != byte_len {
                b.Fatalf("%+v", zerrors.NewSimplef("测试失败, 收到的值长度非预期 %s, %d", k, len(*a)))
            }

            // 超过10000个key不做值检查
            if max_key_count <= 10000 {
                if !bytes.Equal(*a, vs[m]) {
                    b.Fatalf("%+v", zerrors.WrapSimplef(err, "测试失败, 收到的值结果非预期 %s, %T", k, a))
                }
            }
        }
    })
}
