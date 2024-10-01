package global

import (
	ut "github.com/go-playground/universal-translator"
	"mxshop_api/user-web/config"
	"mxshop_api/user-web/proto"
)

var (
	Trans         ut.Translator
	ServerConfig  = &config.ServerConfig{}
	UserSrvClient proto.UserClient
)
