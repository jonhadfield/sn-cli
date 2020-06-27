package sncli

import (
	"github.com/jonhadfield/gosn-v2"
)

type RegisterConfig struct {
	Email     string
	Password  string
	APIServer string
	Debug     bool
}

func (i *RegisterConfig) Run() error {
	registerInput := gosn.RegisterInput{
		Email:     i.Email,
		Password:  i.Password,
		APIServer: i.APIServer,
	}

	_, err := registerInput.Register()

	return err
}
