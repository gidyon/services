package main

import (
	"encoding/base64"
	"fmt"

	"github.com/pjebs/optimus-go"
)

func main() {
	o := optimus.New(1580030173, 59260789, 1163945558) // Prime Number: 1580030173, Mod Inverse: 59260789, Pure Random Number: 1163945558

	new_id := o.Encode(1) // internal id of 15 being transformed to 1103647397

	fmt.Println(new_id)

	orig_id := o.Decode(1103647397)

	fmt.Println(orig_id)

	str := fmt.Sprint(10)

	base64str := base64.StdEncoding.EncodeToString([]byte(str))

	fmt.Println(base64str)

	bs, err := base64.StdEncoding.DecodeString(base64str)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bs))
}
