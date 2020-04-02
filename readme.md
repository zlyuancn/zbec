# 朴实无华的后端缓存, 单实例百万+并发, 解决缓存击穿丶缓存雪崩丶缓存穿透, 支持任何数据库

---

# 获得
`go get -u github.com/zlyuancn/zbec`

# 解决缓存击穿

> 当有多个进程同时获取一个key时, 只有一个进程会真的去缓存db读取或从db加载并返回结果, 其他的进程会等待该进程结束直接收到结果. 实现方式请转到 [github.com/zlyuancn/zsingleflight](https://github.com/zlyuancn/zsingleflight)

# 解决缓存雪崩

+ 设置随机的TTL, 可以有效减小缓存雪崩的风险

# 解决缓存穿透

+ 可以通过 `zbec.WithCacheNilData` 开启缓存空数据(默认开启), 空数据有效时间默认为 5s
+ 可以通过 `zbec.WithLocalCache` 设置本地缓存
+ 在用户请求key的时候判断它是否可能不存在, 比如判断id长度不等于32(uuid去掉横杠的长度)直接返回错误

# db数据库
+ 支持任何数据库, 本模块不关心用户如何加载数据

# 缓存数据库
+ [任何实现 `zbec.ICacheDB` 的结构](cachedb.go)
+ [redis](./cachedb/redis/c.go)
+ [go-cache](./cachedb/go_cache/c.go)

# 编解码器

> 开发过程中不需要考虑每个对象的编解码, 可以在初始化时为缓存数据库时选择一个编解码器, 默认是`MsgPack`

+ [任何实现 `codec.ICodec` 的结构](./codec/codec.go)
+ Byte
+ JSON
+ JsonIterator
+ MsgPack
+ ProtoBuffer
+ Thrift

# 以下是性能测试数据

> 未模拟用户请求和db加载, 直接测试本模块本身的性能

```shell script
# 性能测试命令
go test -v -bench "^Benchmark_.+$" -run ^$ -cpu 20,50,100,1000,10000 .
# 下面这个是用户没有go环境但是有docker的情况下
docker run --rm -v ${PWD}/../..:/src/app -v /src/gopath:/src/gopath -v /src/gocache:/src/gocache -w /src/app/zbec/test zlyuan/golang:1.13 go test -v -bench "^Benchmark_.+$" -run ^$ -cpu 20,50,100,1000,10000 .
```

# 1000 个key, 每个key 512字节随机数据, 请求key顺序随机

```
# go-cache
Benchmark_GoCache1e3-20                     	 1885130	       730 ns/op
Benchmark_GoCache1e3-50                     	 1568814	       777 ns/op
Benchmark_GoCache1e3-100                    	 2048479	       744 ns/op
Benchmark_GoCache1e3-1000                   	 2303812	       642 ns/op
Benchmark_GoCache1e3-10000                  	 1705032	       716 ns/op

# redis
Benchmark_Redis1e3-20                       	   88453	     12643 ns/op
Benchmark_Redis1e3-50                       	  229730	      4961 ns/op
Benchmark_Redis1e3-100                      	  459921	      2482 ns/op
Benchmark_Redis1e3-1000                     	 2120120	       555 ns/op
Benchmark_Redis1e3-10000                    	 1548505	       722 ns/op

# redis and LocalCache
Benchmark_RedisAndLocalCache1e3-20          	 1948473	       704 ns/op
Benchmark_RedisAndLocalCache1e3-50          	 1781548	       697 ns/op
Benchmark_RedisAndLocalCache1e3-100         	 1971524	       697 ns/op
Benchmark_RedisAndLocalCache1e3-1000        	 2467428	       567 ns/op
Benchmark_RedisAndLocalCache1e3-10000       	 2359215	       539 ns/op
```


# 10 000 个key, 每个key 512字节随机数据, 请求key顺序随机

<details>
<summary>点击展开</summary>
<pre><code>
# go-cache
Benchmark_GoCache1e4-20                     	 2310266	       495 ns/op
Benchmark_GoCache1e4-50                     	 1808149	       734 ns/op
Benchmark_GoCache1e4-100                    	 1820456	       646 ns/op
Benchmark_GoCache1e4-1000                   	 2386491	       639 ns/op
Benchmark_GoCache1e4-10000                  	 2307810	       490 ns/op
<br/>
# redis
Benchmark_Redis1e4-20                       	  102188	     11582 ns/op
Benchmark_Redis1e4-50                       	  229939	      4651 ns/op
Benchmark_Redis1e4-100                      	  518737	      2286 ns/op
Benchmark_Redis1e4-1000                     	 2165594	       576 ns/op
Benchmark_Redis1e4-10000                    	 1486389	       753 ns/op
<br/>
# redis and LocalCache
Benchmark_RedisAndLocalCache1e4-20          	  141720	      8330 ns/op
Benchmark_RedisAndLocalCache1e4-50          	  386697	      2975 ns/op
Benchmark_RedisAndLocalCache1e4-100         	  595728	      1722 ns/op
Benchmark_RedisAndLocalCache1e4-1000        	 2459406	       516 ns/op
Benchmark_RedisAndLocalCache1e4-10000       	 2012082	       606 ns/op
</code></pre>
</details>

## 100 000 个key, 每个key 512字节随机数据, 请求key顺序随机

<details>
<summary>点击展开</summary>
<pre><code>
# go-cache
Benchmark_GoCache1e5-20                     	 1325640	       873 ns/op
Benchmark_GoCache1e5-50                     	 2432680	       625 ns/op
Benchmark_GoCache1e5-100                    	 2065634	       737 ns/op
Benchmark_GoCache1e5-1000                   	 2319595	       657 ns/op
Benchmark_GoCache1e5-10000                  	 2056987	       602 ns/op
<br/>
# redis
Benchmark_Redis1e5-20                       	   96511	     12034 ns/op
Benchmark_Redis1e5-50                       	  220322	      5035 ns/op
Benchmark_Redis1e5-100                      	  461415	      2308 ns/op
Benchmark_Redis1e5-1000                     	 2163451	       615 ns/op
Benchmark_Redis1e5-10000                    	 1538402	       710 ns/op
<br/>
# redis and LocalCache
Benchmark_RedisAndLocalCache1e5-20          	   99225	     11235 ns/op
Benchmark_RedisAndLocalCache1e5-50          	  248864	      4999 ns/op
Benchmark_RedisAndLocalCache1e5-100         	  427542	      2480 ns/op
Benchmark_RedisAndLocalCache1e5-1000        	 2517938	       566 ns/op
Benchmark_RedisAndLocalCache1e5-10000       	 1857760	       562 ns/op
</code></pre>
</details>

# 开发注意事项

+ 获取值时保存结果的变量必须是一个指针
+ 从数据库加载结果为空时应该返回`zbec.NilData`错误
+ 本地缓存一定会缓存空数据
+ 缓存数据库会根据设置的空数据过期时间缓存空数据, 默认为 5s
+ 如果你可能对结果进行修改, 为了不产生并发写错误, 你应该设置 `zbec.WithDeepcopyResult(true)` , 代价是性能有所降低

# 示例代码

> 即将出现...
