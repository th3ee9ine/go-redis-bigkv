package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"sort"
	"strconv"
)

var ctx = context.Background()

var (
	addr     string
	password string
	db       int
	maxSize  int64
)

func init() {
	flag.StringVar(&addr, "a", "localhost:3306", "Please input redis addr")
	flag.StringVar(&password, "p", "", "Please input redis password")
	flag.IntVar(&db, "db", 0, "Please input redis db")
	flag.Int64Var(&maxSize, "m", 10240, "Please input big key size(b)")
}

type KeyInfoEntity struct {
	KeyName string
	Size    int64
	KeyType string
}

type KeyInfoEntitys []KeyInfoEntity

// Len 获取此 slice 的长度
func (p KeyInfoEntitys) Len() int {
	return len(p)
}

// Less 比较两个元素大小 降序
func (p KeyInfoEntitys) Less(i, j int) bool {
	return p[i].Size > p[j].Size
}

// Swap 交换数据
func (p KeyInfoEntitys) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p KeyInfoEntity) String() string {
	st := "key的类型: " + p.KeyType + " key的名称: " + p.KeyName + " value的内存大小: "

	//1024Byte（字节）= 1K(千字节)
	//1024K（千字节）= 1M（兆字节）
	//1024M = 1G
	//1024G = 1T
	if 1024<<40 <= p.Size {
		value := p.Size >> 40
		st += strconv.FormatInt(value, 10)
		st += "T"
	} else if 1024<<30 <= p.Size {
		value := p.Size >> 30
		st += strconv.FormatInt(value, 10)
		st += "G"
	} else if 1024<<20 <= p.Size {
		value := p.Size >> 20
		st += strconv.FormatInt(value, 10)
		st += "M"
	} else if 1024<<10 <= p.Size {
		value := p.Size >> 10
		st += strconv.FormatInt(value, 10)
		st += "K"
	} else {
		st += strconv.FormatInt(p.Size, 10)
		st += "Byte"
	}
	if maxSize < p.Size {
		st += " 大key警告，请优化"
	}
	return st
}

// scanBigKey keyType枚举类型为：string、list、set、hash、zset
func scanBigKey(rdb *redis.Client, keyType string, ch chan []KeyInfoEntity) {
	var keyList []KeyInfoEntity
	d := rdb.ScanType(ctx, 0, "", 0, keyType)
	iterator := d.Iterator()
	for iterator.Next(ctx) {
		key := iterator.Val()
		c := rdb.MemoryUsage(ctx, key)
		keyEntity := KeyInfoEntity{
			KeyName: key,
			Size:    c.Val(),
			KeyType: keyType,
		}
		keyList = append(keyList, keyEntity)
	}
	ch <- keyList
}

func main() {
	flag.PrintDefaults()
	//Scans the arg list and sets up flags
	flag.Parse()

	var rdb = redis.NewClient(&redis.Options{
		Addr: addr,
		// redis 密码，没有就为空
		Password: password,
		// 可改成使用的db
		DB:       db,
		PoolFIFO: true,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	ch := make(chan []KeyInfoEntity, 5)
	go scanBigKey(rdb, "string", ch)
	go scanBigKey(rdb, "set", ch)
	go scanBigKey(rdb, "zset", ch)
	go scanBigKey(rdb, "list", ch)
	go scanBigKey(rdb, "hash", ch)

	var keyList []KeyInfoEntity
	for i := 0; i < 5; i++ {
		temp := <-ch
		keyList = append(keyList, temp...)
	}

	sort.Sort(KeyInfoEntitys(keyList))

	outputFile, outputError := os.OpenFile("output.txt", os.O_WRONLY|os.O_CREATE, 0666)

	if outputError != nil {
		fmt.Printf("An error occurred with file opening or creation\n")
		return
	}
	defer func() {
		outputFile.Close()
		rdb.Close()
	}()

	outputWriter := bufio.NewWriter(outputFile)
	for i, _ := range keyList {
		outputWriter.WriteString(keyList[i].String() + "\n")
	}
	outputWriter.Flush()
}
