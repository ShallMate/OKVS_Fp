package main

import (
	rand1 "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	okvs "github.com/OKVSFp/OKVS"
	"github.com/tunabay/go-bitarray"
)

func Testcoding() {
	for {
		value := rand.Uint32()
		Value := bitarray.NewFromInt(big.NewInt(int64(value)))
		Value = Value.ToWidth(32, bitarray.AlignRight)
		value1 := Value.ToInt().Int64()
		if value1 != int64(value) {
			fmt.Println("coding fail")
		}
		fmt.Println(value)
		fmt.Println(value1)
	}

}

func main() {
	n := 1000 // 假设长度为 5
	m := int(math.Round(float64(n) * 1.03))

	// 创建长度为 n 的 KV 结构体切片
	kvs := make([]okvs.KV, n)
	q, _ := rand1.Prime(rand1.Reader, 32)
	// 输出 KV 结构体切片
	for i := 0; i < n; i++ {
		key, _ := rand1.Int(rand1.Reader, q)     // 生成长度为8的随机字节切片作为key
		value, _ := rand1.Int(rand1.Reader, q)   // 生成随机的uint32切片作为value
		kvs[i] = okvs.KV{Key: key, Value: value} // 将key和value赋值给KV结构体
	}
	//fmt.Printf("KV slice: %+v\n", kvs)
	okvs := okvs.OKVS{
		N: n,                   // 假设 N 的长度为 10
		M: m,                   // 假设 M 的长度为 10
		W: 360,                 // 假设 W 的长度为 32
		P: make([]*big.Int, m), // 初始化 P 切片长度为 10
		Q: q,
	}
	okvs.Encode(kvs)
	for i := 0; i < int(n); i++ {
		v := okvs.Decode(kvs[i].Key)
		fmt.Println(v)
		fmt.Println(kvs[i].Value)
	}

}
