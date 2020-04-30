package mail

import (
	"fmt"
	"time"

	"webcron/app/libs"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/utils"

)

var (
	sendCh          chan *utils.Email
	config          string
	EMailAttachSize int64
)

func init() {
	queueSize, _ := beego.AppConfig.Int("mail.queue_size")
	host := beego.AppConfig.String("mail.host")
	port, _ := beego.AppConfig.Int("mail.port")
	username := beego.AppConfig.String("mail.user")
	password := beego.AppConfig.String("mail.password")
	from := beego.AppConfig.String("mail.from")
	EMailAttachSize, _ = beego.AppConfig.Int64("mail.attachsize")
	EMailAttachSize = EMailAttachSize << 20 // EMailAttachSize M
	if port == 0 {
		port = 25
	}

	config = fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","port":%d,"from":"%s"}`, username, password, host, port, from)

	sendCh = make(chan *utils.Email, queueSize)

	go func() {
		for {
			select {
			case m, ok := <-sendCh:
				if !ok {
					return
				}
				if err := m.Send(); err != nil {
					beego.Error("SendMail:", err.Error())
				}
			}
		}
	}()
}

func SendMail(address, name, subject, content string, cc []string) bool {
	mail := utils.NewEMail(config)
	mail.To = []string{address}
	mail.Subject = subject
	mail.HTML = content
	if len(cc) > 0 {
		mail.Cc = cc
	}

	select {
	case sendCh <- mail:
		return true
	case <-time.After(time.Second * 3):
		return false
	}
}

func SendMailWithAttach(subject, content, attach string, cc []string) bool {
	if len(cc) == 0 {
		return false
	}
	mail := utils.NewEMail(config)
	mail.To = []string{cc[0]}
	mail.Subject = subject
	mail.HTML = content
	if len(attach) > 0 && libs.IsPathExist(attach) {
		_, err := mail.AttachFile(attach)
		if err != nil {
			return false
		}
	}
	ccTmp := cc[1:]
	if len(ccTmp) > 0 {
		mail.Cc = ccTmp
	}
	select {
	case sendCh <- mail:
		return true
	case <-time.After(time.Second * 3):
		return false
	}
}
