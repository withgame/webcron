package controllers

import (
	"github.com/astaxie/beego"
)

var (
	basicAuthName string
	basicAuthPwd  string
	copyLink      = "http://www.lisijie.org"
	copyName      = "lisijie.org"
	copyStartYear = "2015"
)

func Init() {
	basicAuthName = beego.AppConfig.String("auth.username")
	basicAuthPwd = beego.AppConfig.String("auth.password")
	if len(beego.AppConfig.String("copy.link")) > 0 {
		copyLink = beego.AppConfig.String("copy.link")
	}
	if len(beego.AppConfig.String("copy.name")) > 0 {
		copyName = beego.AppConfig.String("copy.name")
	}
	if len(beego.AppConfig.String("copy.start")) > 0 {
		copyStartYear = beego.AppConfig.String("copy.start")
	}
}
