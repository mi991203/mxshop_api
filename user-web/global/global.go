package global

import (
	ut "github.com/go-playground/universal-translator"

	"mxshop_api/user-web/config"
)

var (
	Trans        ut.Translator
	ServerConfig = &config.ServerConfig{}
)
