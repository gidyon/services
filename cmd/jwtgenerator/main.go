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
	jwtKey = flag.String("jwt-key", "", "Key to use for signing JWT")
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
	}, time.Now().Add(100*24*time.Hour))
	errs.Panic(err)

	fmt.Println(token)
}
