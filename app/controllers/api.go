package controllers

import (
	"crypto/subtle"
	"strings"

	libcron "github.com/lisijie/cron"
	"webcron/app/libs"
	"webcron/app/mail"
	"webcron/app/models"
)

type ApiController struct {
	BaseController
}

func (this *ApiController) Create() {
	basicUserName, basicPassword, ok := this.Ctx.Request.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(basicUserName), []byte(basicAuthName)) != 1 || subtle.ConstantTimeCompare([]byte(basicPassword), []byte(basicAuthPwd)) != 1 {
		this.Ctx.Request.Header.Set("WWW-Authenticate", `Basic realm="`+this.controllerName+`"`)
		this.Ctx.ResponseWriter.WriteHeader(401)
		this.Ctx.ResponseWriter.Write([]byte("Unauthorised.\n"))
		return
	}
	if !this.isPost() {
		this.StopRun()
		return
	}
	task := new(models.Task)
	task.UserId = 0
	task.GroupId, _ = this.GetInt("gid")
	task.TaskName = strings.TrimSpace(this.GetString("tname"))
	//task.Description = strings.TrimSpace(this.GetString("tdesc"))
	//task.Concurrent, _ = this.GetInt("concurrent")
	task.Concurrent = 0
	task.CronSpec = strings.TrimSpace(this.GetString("spec"))
	task.Command = strings.TrimSpace(this.GetString("command"))
	task.Notify, _ = this.GetInt("notify")
	task.Timeout, _ = this.GetInt("timeout")
	task.TotalTimes, _ = this.GetInt("total")
	task.TaskType = models.TaskType_HTTP
	notifyEmail := strings.TrimSpace(this.GetString("notify_email"))
	if notifyEmail != "" {
		emailList := make([]string, 0)
		tmp := strings.Split(notifyEmail, "\n")
		for _, v := range tmp {
			v = strings.TrimSpace(v)
			if !libs.IsEmail([]byte(v)) {
				this.ajaxMsg("无效的Email地址："+v, MSG_ERR)
			} else {
				emailList = append(emailList, v)
			}
		}
		task.NotifyEmail = strings.Join(emailList, "\n")
	}
	notifyEmailAttach := strings.TrimSpace(this.GetString("notify_email_attach"))
	if notifyEmailAttach != "" {
		task.NotifyEmailAttach = notifyEmailAttach
	}
	if task.TaskName == "" || task.CronSpec == "" || task.Command == "" {
		this.ajaxMsg("请填写完整信息", MSG_ERR)
	}
	if _, err := libcron.Parse(task.CronSpec); err != nil {
		this.ajaxMsg("cron表达式无效", MSG_ERR)
	}
	if _, err := models.TaskAdd(task); err != nil {
		this.ajaxMsg(err.Error(), MSG_ERR)
	}
	executeImmediately(task)
	this.ajaxMsg("", MSG_OK)
	return
}

func (this *ApiController) Mail() {
	basicUserName, basicPassword, ok := this.Ctx.Request.BasicAuth()
	if !ok || subtle.ConstantTimeCompare([]byte(basicUserName), []byte(basicAuthName)) != 1 || subtle.ConstantTimeCompare([]byte(basicPassword), []byte(basicAuthPwd)) != 1 {
		this.Ctx.Request.Header.Set("WWW-Authenticate", `Basic realm="`+this.controllerName+`"`)
		this.Ctx.ResponseWriter.WriteHeader(401)
		this.Ctx.ResponseWriter.Write([]byte("Unauthorised.\n"))
		return
	}
	if !this.isPost() {
		this.StopRun()
		return
	}
	subject := this.GetString("subject")
	content := this.GetString("content")
	ems := this.GetStrings("ems")
	if len(subject) == 0 {
		this.ajaxMsg("missing subject", MSG_ERR)
	}
	if len(content) == 0 {
		this.ajaxMsg("missing content", MSG_ERR)
	}
	if len(ems) == 0 {
		this.ajaxMsg("missing ems", MSG_ERR)
	}
	var ccEms []string
	if len(ems) > 1 {
		ccEms = ems[1:]
	}
	if !mail.SendMail(ems[0], "", subject, content, ccEms) {
		this.ajaxMsg("mail send err", MSG_ERR)
	}
	this.ajaxMsg("ok", MSG_OK)
	return
}
