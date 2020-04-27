/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/27
   Description :
-------------------------------------------------
*/

package main

import (
    "bytes"
    "flag"
    "fmt"
    "log"
    "math/rand"
    "time"

    rredis "github.com/go-redis/redis"

    "github.com/zlyuancn/zbec"
    "github.com/zlyuancn/zbec/cachedb/redis"
    "github.com/zlyuancn/zbec/codec"
    "github.com/zlyuancn/zbec/query"
)

func failOnErr(err error, msg string) {
    if err != nil {
        log.Fatalf("%s: %s", msg, err.Error())
    }
}

func getRedisClient(on_local_cache bool, host string, pwd string, db int) *zbec.BECache {
    cdb := rredis.NewClient(&rredis.Options{
        Addr:        host,
        Password:    pwd,
        DB:          db,
        PoolSize:    10,
        DialTimeout: time.Second * 3,
    })
    if err := cdb.Ping().Err(); err != nil {
        panic(err)
    }

    return zbec.New(
        redis.Wrap(
            cdb,
            redis.WithCodecType(codec.Byte),
        ),
        zbec.WithLocalCache(on_local_cache),
    )
}

func getGoCache() *zbec.BECache {
    return zbec.NewOfGoCache(0)
}

func benchmark_any(bec *zbec.BECache, max_key_count int, client_num int, second int) {
    rand.Seed(time.Now().UnixNano())

    byte_len := 512

    space := "benchmark"
    loader := zbec.NewNameLoader(space, func(query *query.Query) (interface{}, error) {
        bs := make([]byte, byte_len)
        for i := 0; i < len(bs); i++ {
            bs[i] = byte(rand.Int() % 256)
        }
        return &bs, nil
    })
    bec.RegisterLoader(loader)

    log.Print("初始化数据")
    vs := make([][]byte, max_key_count)
    {
        for i := 0; i < max_key_count; i++ {
            k := fmt.Sprintf("%d", i)

            a := new([]byte)
            err := bec.Get(zbec.NewQuery(space, k, ""), a)
            if err != nil {
                failOnErr(err, "初始化失败")
            }

            if len(*a) != byte_len {
                log.Fatal("初始化失败, 收到的值长度非预期")
            }
            vs[i] = *a
        }
    }

    log.Print("开始")

    done := make(chan struct{})
    go func() {
        time.Sleep(time.Second * time.Duration(second))
        close(done)
    }()

    tn := time.NewTicker(time.Second)
    run := true
    for run {
        select {
        case <-done:
            tn.Stop()
            run = false
            break
        case <-tn.C:
            for j := 0; j < client_num; j++ {
                go func() {
                    rn := rand.Int()
                    time.Sleep(time.Duration(rn%1000+1) * time.Millisecond) // 随机等待1秒内

                    m := rn % max_key_count
                    k := fmt.Sprintf("%d", m)

                    a := new([]byte)
                    err := bec.Get(zbec.NewQuery(space, k, ""), a)
                    failOnErr(err, "测试失败")
                    if len(*a) != byte_len {
                        log.Fatal("测试失败, 收到的值长度非预期")
                    }

                    // 超过10000个key不做值检查
                    if !bytes.Equal(*a, vs[m]) {
                        log.Fatal("测试失败, 收到的值长度非预期")
                    }
                }()
            }
        }
    }

    log.Print("结束")
}

func main() {
    bec_type := flag.String("bec", "gocache", "bec类型")
    key_num := flag.Int("key_num", 1000, "key数量")
    client_num := flag.Int("client_num", 1000, "客户端数量")
    second_num := flag.Int("second_num", 5, "执行时间(s)")
    redis_host := flag.String("redis_host", "127.0.0.1:6379", "redis地址")
    redis_pwd := flag.String("redis_pwd", "", "redis密码")
    redis_db := flag.Int("redis_db", 0, "redis的db")
    flag.Parse()

    var bec *zbec.BECache
    switch *bec_type {
    case "gocache":
        bec = getGoCache()
    case "redis":
        bec = getRedisClient(false, *redis_host, *redis_pwd, *redis_db)
    case "redis_and_localcache":
        bec = getRedisClient(true, *redis_host, *redis_pwd, *redis_db)
    default:
        log.Fatal("bec类型为gocache,redis,redis_and_localcache")
    }
    benchmark_any(bec, *key_num, *client_num, *second_num)
}
