package main

import (
	"testing"

	"github.com/Pallinder/go-randomdata"
	"github.com/gidyon/micro/v2/utils/encryption"
	"github.com/speps/go-hashids"
)

// NewHasher creates a new hasher for decoding and encoding int64 slices
func NewHasher(salt string) (*hashids.HashID, error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 30
	return hashids.NewWithData(hd)
}

func BenchmarkHasher(b *testing.B) {
	hasher, err := NewHasher(string(randomdata.RandStringRunes(32)))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		i64 := int64(1)
		hash, err := hasher.EncodeInt64([]int64{int64(i64)})
		if err != nil {
			b.Fatal(err)
		}
		v := hasher.DecodeInt64(hash)
		if len(v) != 1 {
			b.Fatalf("expected length 1 got: %d", len(v))
		}
		if v[0] != i64 {
			b.Fatalf("expected %d == %d to be true", v[0], i64)
		}
	}
}

func BenchmarkHasherEncode(b *testing.B) {
	hasher, err := encryption.NewHasher(string(randomdata.RandStringRunes(32)))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		i64 := int64(1)
		hash, err := hasher.EncodeInt64([]int64{int64(i64)})
		if err != nil {
			b.Fatal(err)
		}
		_ = hash
	}
}
