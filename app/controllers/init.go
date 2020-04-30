package controllers

import (
	"github.com/astaxie/beego"
)

var (
	basicAuthName string
	basicAuthPwd  string
)

func Init() {
	basicAuthName = beego.AppConfig.String("auth.username")
	basicAuthPwd = beego.AppConfig.String("auth.password")
}
