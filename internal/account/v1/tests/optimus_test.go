package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"testing"

	"github.com/pjebs/optimus-go"
)

func BenchmarkOptimus(b *testing.B) {
	o := optimus.New(1580030173, 59260789, 1163945558)

	for i := 0; i < b.N; i++ {
		ui64 := uint64(1)
		val := o.Encode(ui64)
		v := o.Decode(val)
		if v != ui64 {
			b.Fatalf("expected %d == %d to be true", v, ui64)
		}
	}
}

func BenchmarkBase64Encode(b *testing.B) {

	for i := 0; i < b.N; i++ {
		str := fmt.Sprint(i)
		base64str := base64.StdEncoding.EncodeToString([]byte(str))
		bs, err := base64.StdEncoding.DecodeString(base64str)
		if err != nil {
			b.Fatal(err)
		}
		if string(bs) != str {
			b.Fatalf("expected %s == %s to be true", string(bs), str)
		}
	}
}
func BenchmarkBase64EncodeStrconv(b *testing.B) {

	for i := 0; i < b.N; i++ {
		str := fmt.Sprint(i)
		base64str := base64.StdEncoding.EncodeToString([]byte(str))
		bs, err := base64.StdEncoding.DecodeString(base64str)
		if err != nil {
			b.Fatal(err)
		}
		str2 := string(bs)
		if str2 != str {
			b.Fatalf("expected %s == %s to be true", str2, str)
		}
		v, err := strconv.ParseUint(str2, 10, 32)
		if err != nil {
			b.Fatalf("expected err to nil got: %v", err)
		}
		_ = v
	}
}

func BenchmarkOptimusEncodeV3(b *testing.B) {
	o := optimus.New(1580030173, 59260789, 1163945558)

	for i := 0; i < b.N; i++ {
		ui64 := uint64(i)
		val := o.Encode(ui64)
		str := fmt.Sprint(val)
		base64str := base64.StdEncoding.EncodeToString([]byte(str))
		bs, err := base64.StdEncoding.DecodeString(base64str)
		if err != nil {
			b.Fatal(err)
		}
		if string(bs) != str {
			b.Fatalf("expected %s == %s to be true", string(bs), str)
		}

	}
}
