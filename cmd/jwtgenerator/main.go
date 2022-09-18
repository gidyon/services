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
	projectId     = flag.String("project-id", "xxxx", "Project id")
	audience      = flag.String("audience", "accounts", "JWT Issuer audience")
	actorId       = flag.String("actor-id", "xxxx", "Actor id")
	names         = flag.String("names", "xxxx", "JWT names")
	iss           = flag.String("iss", "xxxx", "JWT names")
	expireSeconds = flag.Int("expire-sec", (60 * 60 * 24 * 365 * 5), "Time in seconds when jwt should expire")
)

func main() {
	flag.Parse()

	jwtKey2 := []byte(*jwtKey)

	// Authentication API
	authAPI, err := auth.NewAPI(&auth.Options{
		SigningKey: jwtKey2,
		Issuer:     *iss,
		Audience:   *audience,
	})
	errs.Panic(err)

	// Generate jwt token
	token, err := authAPI.GenToken(context.Background(), &auth.Payload{
		ID:           *actorId,
		ProjectID:    *projectId,
		Names:        *names,
		PhoneNumber:  "xxxxxx",
		EmailAddress: "xxxxxx",
		Group:        auth.DefaultAdminGroup(),
	}, time.Now().Add(time.Second*time.Duration(*expireSeconds)))
	errs.Panic(err)

	fmt.Println(token)
}
