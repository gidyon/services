package mocks

import (
	"github.com/gidyon/services/pkg/mocks/mocks"
	"github.com/stretchr/testify/mock"
)

// Cooky is generic cookie encoder and decorder
type Cooky interface {
	Decode(string, string, interface{}) error
	Encode(string, interface{}) (string, error)
}

// Cookier is an instance of Cooky
var Cookier = &mocks.Cooky{}

func init() {
	Cookier.On("Encode", mock.Anything, mock.Anything).
		Return(mock.Anything, nil)

	Cookier.On("Decode", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
}
