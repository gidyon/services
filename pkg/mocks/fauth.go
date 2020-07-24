package mocks

import (
	"context"

	fauth "firebase.google.com/go/auth"
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/stretchr/testify/mock"
)

// FirebaseAuth contains methods used in firebase authentication
type FirebaseAuth interface {
	VerifyIDToken(context.Context, string) (*fauth.Token, error)
}

// FirebaseAuthAPI is firebase auth mocked
var FirebaseAuthAPI = &mocks.FirebaseAuth{}

func init() {
	FirebaseAuthAPI.On("VerifyIDToken", mock.Anything, mock.Anything).
		Return(&fauth.Token{}, nil)
}
