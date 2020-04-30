/*
* File:    init.go
* Created: 2020-04-30 16:07
* Authors: MS geek.snail@qq.com
* Copyright (c) 2013 - 2020 虾游网络科技有限公司 版权所有
 */

package libs

import (
	"github.com/astaxie/beego"
)

var (
	webCronQnHub          string
	webCronQnPrivateHub   string
	webCronQnAK           string
	webCronQnSK           string
	webCronQnZone         string
	qnAccessDomain        string
	qnPrivateAccessDomain string
)

func Init() {
	qnAccessDomain = beego.AppConfig.String("qn.domain")
	qnPrivateAccessDomain = beego.AppConfig.String("qn.pivatedomain")
	webCronQnAK = beego.AppConfig.String("qn.ak")
	webCronQnSK = beego.AppConfig.String("qn.sk")
	webCronQnHub = beego.AppConfig.String("qn.hub")
	webCronQnZone = beego.AppConfig.String("qn.zone")
	webCronQnPrivateHub = beego.AppConfig.String("qn.privatehub")
}
