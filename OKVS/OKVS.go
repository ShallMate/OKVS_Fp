package okvs

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"

	"github.com/tunabay/go-bitarray"
	"golang.org/x/crypto/blake2b"
)

// 定义System结构体
type System struct {
	Pos   int
	Row   []*big.Int
	Value *big.Int
}

/*
docstring for OKVS
:n: rows
:m: columns
:w: length of band
The choice of parameters:
m = (1 + epsilon)n
w = O(lambda / epsilon + log n)
For example:
m = 2^10, epsilon = 0.1,
==> n = (1+0.1) * 2^10
*/

var one = big.NewInt(1)
var zero = big.NewInt(0)

type OKVS struct {
	N int //okvs存储的k-v长度
	M int //okvs的实际长度
	W int //随机块的长度
	P []*big.Int
	Q *big.Int
}

type KV struct {
	Key   *big.Int //key
	Value *big.Int //value
}

func HashToFixedSize(bytesize int, key *big.Int) []byte {
	//fmt.Println(bytesize)
	hash, _ := blake2b.New(bytesize, key.Bytes())
	return hash.Sum(nil)[:]
}

func (r *OKVS) hash1(bytesize int, key *big.Int) int {
	hashRange := r.M - r.W
	hashkey := HashToFixedSize(bytesize, key)
	hashkeyint := int(binary.BigEndian.Uint32(hashkey)) % hashRange
	return hashkeyint
}

func (r *OKVS) hash2(pos int, key *big.Int) *bitarray.BitArray {
	bandsize := int(r.W) / 8
	hashBytes := HashToFixedSize(bandsize, key)
	band := bitarray.NewFromBytes(hashBytes, 0, r.W)
	band = band.ToWidth(r.W+pos, bitarray.AlignRight)
	band = band.ToWidth(r.M, bitarray.AlignLeft)
	return band
}

func (r *OKVS) Init(kvs []KV) []System {
	systems := make([]System, r.N)
	for i := 0; i < r.N; i++ {
		systems[i].Pos = r.hash1(4, kvs[i].Key)
		systems[i].Row = make([]*big.Int, r.M)
		row := r.hash2(systems[i].Pos, kvs[i].Key)
		for j := 0; j < r.M; j++ {
			if row.BitAt(j) == 1 {
				systems[i].Row[j] = one
			} else {
				systems[i].Row[j] = zero
			}
		}
		//fmt.Println(systems[i].Row)
		systems[i].Value = kvs[i].Value
	}
	for i := 0; i < r.M; i++ {
		r.P[i] = zero
	}
	return systems
}

func (r *OKVS) Encode(kvs []KV) *OKVS {
	if len(kvs) != r.N {
		fmt.Println("r.N must equal to len(kvs)")
		return nil
	}
	systems := r.Init(kvs)
	//fmt.Println(systems)
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].Pos < systems[j].Pos
	})
	piv := make([]int, r.N)
	for i := range piv {
		piv[i] = -1
	}
	for i := 0; i < r.N; i++ {
		for j := systems[i].Pos; j < r.W+systems[i].Pos; j++ {
			//fmt.Println(systems[i])
			if systems[i].Row[j].Cmp(zero) != 0 {
				piv[i] = j
				for k := i + 1; k < r.N; k++ {
					if systems[k].Pos <= piv[i] && systems[k].Row[piv[i]].Cmp(zero) != 0 {
						t := new(big.Int).ModInverse(systems[i].Row[j], r.Q)
						t = new(big.Int).Mul(t, systems[k].Row[piv[i]])
						t = new(big.Int).Mod(t, r.Q)
						for s := 0; s < r.M; s++ {
							systems[k].Row[s] = new(big.Int).Sub(systems[k].Row[s], new(big.Int).Mul(t, systems[i].Row[s]))
							systems[k].Row[s] = new(big.Int).Mod(systems[k].Row[s], r.Q)
						}
						systems[k].Value = new(big.Int).Sub(systems[k].Value, new(big.Int).Mul(t, systems[i].Value))
						systems[k].Value = new(big.Int).Mod(systems[k].Value, r.Q)
					}
				}
				break
			}
		}
		if piv[i] == -1 {
			fmt.Println("Fail to generate at {i}th row!", i)
			return nil
		}
	}
	for i := r.N - 1; i >= 0; i-- {
		//reszeroBytes := make([]byte, 4)
		res := big.NewInt(0)
		for j := 0; j < r.M; j++ {
			if systems[i].Row[j].Cmp(zero) != 0 {
				res = new(big.Int).Add(res, new(big.Int).Mul(systems[i].Row[j], r.P[j]))
			}
		}
		res = res.Sub(systems[i].Value, res)
		t := new(big.Int).ModInverse(systems[i].Row[piv[i]], r.Q)
		res = new(big.Int).Mul(t, res)
		r.P[piv[i]] = new(big.Int).Mod(res, r.Q)
	}
	return r
}

func (r *OKVS) Decode(key *big.Int) *big.Int {
	pos := r.hash1(4, key)
	row := r.hash2(pos, key)
	res := big.NewInt(0)
	for j := pos; j < r.W+pos; j++ {
		if row.BitAt(j) == 1 {
			//r.P[j] = r.P[j].ToWidth(32, bitarray.AlignRight)
			res = new(big.Int).Add(res, r.P[j])
		}
	}
	res = new(big.Int).Mod(res, r.Q)
	return res

}
