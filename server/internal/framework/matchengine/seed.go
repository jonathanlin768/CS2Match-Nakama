package matchengine

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
)

func deriveSeed(parts ...interface{}) int64 {
	h := fnv.New64a()
	var buf [8]byte
	for _, part := range parts {
		switch v := part.(type) {
		case string:
			_, _ = h.Write([]byte(v))
		case int:
			binary.LittleEndian.PutUint64(buf[:], uint64(v))
			_, _ = h.Write(buf[:])
		case int64:
			binary.LittleEndian.PutUint64(buf[:], uint64(v))
			_, _ = h.Write(buf[:])
		default:
			_, _ = h.Write([]byte(fmt.Sprintf("%v", v)))
		}
		_, _ = h.Write([]byte{0})
	}
	return int64(h.Sum64() & 0x7fffffffffffffff)
}

func newDerivedRand(parts ...interface{}) *rand.Rand {
	return rand.New(rand.NewSource(deriveSeed(parts...)))
}
