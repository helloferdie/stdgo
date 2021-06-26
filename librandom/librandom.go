package librandom

import (
	cryptorand "crypto/rand"
	"io"
	"math/rand"
	"time"

	"github.com/helloferdie/stdgo/libslice"
)

// GenerateRandInt64 - Get random int64 from range min - max
func GenerateRandInt64(min int64, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min+1) + min
}

// GenerateRandInt64Slice - Get random []int64 from provided slice
func GenerateRandInt64Slice(list []int64) []int64 {
	l := int64(len(list))
	o := []int64{}
	ol := GenerateRandInt64(1, l)

	var i int64
	for i = 0; i < ol; i++ {
		index := GenerateRandInt64(1, l) - 1
		v := list[index]
		_, exist := libslice.ContainsInt64(v, o)
		if !exist {
			o = append(o, list[index])
		}
	}
	return o
}

// GenerateRandNumber -
func GenerateRandNumber(max int) string {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, max)
	n, err := io.ReadAtLeast(cryptorand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}
