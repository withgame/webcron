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
