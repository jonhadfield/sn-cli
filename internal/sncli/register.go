package sncli

import "github.com/jonhadfield/gosn-v2/auth"

type RegisterConfig struct {
	Email     string
	Password  string
	APIServer string
	Debug     bool
}

func (i *RegisterConfig) Run() error {
	registerInput := auth.RegisterInput{
		Email:     i.Email,
		Password:  i.Password,
		APIServer: i.APIServer,
		Debug:     i.Debug,
	}

	_, err := registerInput.Register()

	return err
}
