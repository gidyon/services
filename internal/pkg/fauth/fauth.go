package fauth

import (
	"context"

	fauth "firebase.google.com/go/auth"
)

// FirebaseAuth is auth client for firebase
type FirebaseAuth interface {
	Auth(context.Context) (FirebaseAuthClient, error)
}

// FirebaseAuthClient is firebase auth client
type FirebaseAuthClient interface {
	VerifyIDToken(context.Context, string) (*fauth.Token, error)
}
