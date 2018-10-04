package sncli

import (
	"github.com/jonhadfield/gosn"
)

type RegisterConfig struct {
	Email     string
	Password  string
	APIServer string
	Debug     bool
}

func (input *RegisterConfig) Run() error {
	//gosn.SetErrorLogger(log.Println)
	//if input.Debug {
	//	gosn.SetDebugLogger(log.Println)
	//}
	registerInput := gosn.RegisterInput{
		Email:     input.Email,
		Password:  input.Password,
		APIServer: input.APIServer,
	}
	_, err := registerInput.Register()
	return err
}
