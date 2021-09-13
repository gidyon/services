package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gidyon/micro/v2/pkg/middleware/grpc/auth"
	"github.com/gidyon/micro/v2/utils/errs"
)

var (
	jwtKey        = flag.String("jwt-key", "", "Key to use for signing JWT")
	expireSeconds = flag.Int("expire-sec", (60 * 60 * 24 * 365 * 5), "Time in seconds when jwt should expire")
)

func main() {
	flag.Parse()

	jwtKey2 := []byte(*jwtKey)

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey2,
		Issuer:     "Accounts API",
		Audience:   "accounts",
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{
		ID:           "xxxxxx",
		ProjectID:    "xxxxxx",
		Names:        "xxxxxx",
		PhoneNumber:  "xxxxxx",
		EmailAddress: "xxxxxx",
		Group:        auth.DefaultAdminGroup(),
	}, time.Now().Add(time.Second*time.Duration(*expireSeconds)))
	errs.Panic(err)

	fmt.Println(token)
}
