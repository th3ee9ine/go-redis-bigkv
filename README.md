# go-redis-bigkv（排查 redis 大 key 小工具）

主要是 **基于 memory 命令**，扫描 redis 中所有的 key，并将结果按照 **内存大小进行排序**，并将排序后的 **结果输出到 txt 文件中**。

因为是 **以 scan 延迟计算的方式扫描所有 key**，因此执行过程中不会阻塞 redis，但实例存在大量的 keys 时，命令执行的时间会很长。

# 如何使用

选取对应的平台的程序，然后运行 bigkv 即可。

-a string (Please input redis addr (default "localhost:3306"))

-db int (Please input redis db)

-m int (Please input big key size(b) (default 10240))

-p string (Please input redis password)

运行完后，会生成一个 output.txt 文件，所有结果在里面，如下面这个例子所示:

```
key的类型: list key的名称: list1 value的内存大小: 144Byte
key的类型: string key的名称: key1 value的内存大小: 112Byte
key的类型: hash key的名称: hash1 value的内存大小: 88Byte
key的类型: zset key的名称: zset value的内存大小: 72Byte
key的类型: set key的名称: set value的内存大小: 64Byte
```

注意：目前程序只支持单 db 扫描，且支持 string、list、hash、zset、set 5种类型。
