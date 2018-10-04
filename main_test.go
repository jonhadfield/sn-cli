package sncli

import (
	"os"

	"github.com/jonhadfield/gosn"
)

var (
	sInput = gosn.SignInInput{
		Email:     os.Getenv("SN_EMAIL"),
		Password:  os.Getenv("SN_PASSWORD"),
		APIServer: os.Getenv("SN_SERVER"),
	}
)
