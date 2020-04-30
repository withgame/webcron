package jobs

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"webcron/app/libs"
	"webcron/app/mail"
	"webcron/app/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
)

var (
	mailTpl              *template.Template
	mailAttachTpl        *template.Template
	mailAttachWihtPwdTpl *template.Template
)

func init() {
	mailTpl, _ = template.New("mail_tpl").Parse(`
	你好 {{.username}}，<br/>

<p>以下是任务执行结果：</p>

<p>
任务ID：{{.task_id}}<br/>
任务名称：{{.task_name}}<br/>       
执行时间：{{.start_time}}<br />
执行耗时：{{.process_time}}秒<br />
执行状态：{{.status}}
</p>
<p>-------------以下是任务执行输出-------------</p>
<p>{{.output}}</p>
<p>
--------------------------------------------<br />
本邮件由系统自动发出，请勿回复<br />
如果要取消邮件通知，请登录到系统进行设置<br />
</p>
`)

	mailAttachTpl, _ = template.New("mail_tpl").Parse(`
	你好 {{.username}}，<br/>

<p>以下是任务执行结果：</p>

<p>
任务ID：{{.task_id}}<br/>
任务名称：{{.task_name}}<br/>       
执行时间：{{.start_time}}<br />
执行耗时：{{.process_time}}秒<br />
执行状态：{{.status}}<br />
备注：{{.brief}}<br />
附件有效: 30分钟<br />
附件地址: {{.attach}}
</p>
<p>-------------以下是任务执行输出-------------</p>
<p>{{.output}}</p>
<p>
--------------------------------------------<br />
本邮件由系统自动发出，请勿回复<br />
如果要取消邮件通知，请登录到系统进行设置<br />
</p>
`)

	mailAttachWihtPwdTpl, _ = template.New("mail_tpl").Parse(`
	你好 {{.username}}，<br/>

<p>以下是任务执行结果：</p>

<p>
任务ID：{{.task_id}}<br/>
任务名称：{{.task_name}}<br/>       
执行时间：{{.start_time}}<br />
执行耗时：{{.process_time}}秒<br />
执行状态：{{.status}}<br />
备注：{{.brief}}<br />
解压密码: {{.pwd}}<br />
附件地址: {{.attach}}
</p>
<p>-------------以下是任务执行输出-------------</p>
<p>{{.output}}</p>
<p>
--------------------------------------------<br />
本邮件由系统自动发出，请勿回复<br />
如果要取消邮件通知，请登录到系统进行设置<br />
</p>
`)

}

type Job struct {
	id         int                                               // 任务ID
	logId      int64                                             // 日志记录ID
	name       string                                            // 任务名称
	task       *models.Task                                      // 任务对象
	runFunc    func(time.Duration) (string, string, error, bool) // 执行函数
	status     int                                               // 任务状态，大于0表示正在执行中
	Concurrent bool                                              // 同一个任务是否允许并行执行
}

func NewJobFromTask(task *models.Task) (*Job, error) {
	if task.Id < 1 {
		return nil, fmt.Errorf("ToJob: 缺少id")
	}
	//job := NewCommandJob(task.Id, task.TaskName, task.Command)
	job := NewCommandJobFromTask(task)
	job.task = task
	job.Concurrent = task.Concurrent == 1
	return job, nil
}

func NewCommandJob(id int, name string, command string) *Job {
	job := &Job{
		id:   id,
		name: name,
	}
	job.runFunc = func(timeout time.Duration) (string, string, error, bool) {
		bufOut := new(bytes.Buffer)
		bufErr := new(bytes.Buffer)
		cmd := exec.Command("/bin/bash", "-c", command)
		cmd.Stdout = bufOut
		cmd.Stderr = bufErr
		cmd.Start()
		err, isTimeout := runCmdWithTimeout(cmd, timeout)

		return bufOut.String(), bufErr.String(), err, isTimeout
	}
	return job
}

func NewCommandJobFromTask(task *models.Task) *Job {
	job := &Job{
		id:   task.Id,
		name: task.TaskName,
	}
	if task.TaskType == models.TaskType_CMD {
		job.runFunc = func(timeout time.Duration) (string, string, error, bool) {
			cmdCleanStr := strings.Replace(task.Command, "\r", "", -1)
			cmdSplitArr := strings.Split(cmdCleanStr, "\n")
			if len(cmdSplitArr) > 1 {
				var (
					isTimeout bool
					outStrs   = make([]string, 0, 0)
					errStrs   = make([]string, 0, 0)
					errorsStr = make([]string, 0, 0)
				)
				for _, cmdStr := range cmdSplitArr {
					if len(cmdStr) == 0 {
						continue
					}
					bufOut := new(bytes.Buffer)
					bufErr := new(bytes.Buffer)
					cmd := exec.Command("/bin/bash", "-c", cmdStr)
					cmd.Stdout = bufOut
					cmd.Stderr = bufErr
					err := cmd.Start()
					if err != nil {
						errorsStr = append(errorsStr, err.Error())
						continue
					}
					err, isTimeout = runCmdWithTimeout(cmd, timeout)
					errorsStr = append(errorsStr, err.Error())
					outStrs = append(outStrs, bufOut.String())
					errStrs = append(errStrs, bufErr.String())
				}
				return strings.Join(outStrs, ","), strings.Join(errStrs, ","), errors.New(strings.Join(errorsStr, ",")), isTimeout
			} else {
				bufOut := new(bytes.Buffer)
				bufErr := new(bytes.Buffer)
				cmd := exec.Command("/bin/bash", "-c", task.Command)
				cmd.Stdout = bufOut
				cmd.Stderr = bufErr
				err := cmd.Start()
				if err != nil {
					return "", "", err, false
				}
				err, isTimeout := runCmdWithTimeout(cmd, timeout)
				return bufOut.String(), bufErr.String(), err, isTimeout
			}
		}
	} else if task.TaskType == models.TaskType_HTTP {
		job.runFunc = func(timeout time.Duration) (string, string, error, bool) {
			commandUri, _ := url.Parse(task.Command)
			params := commandUri.Query()
			notifyUrlStr := fmt.Sprintf("%s://%s%s", commandUri.Scheme, commandUri.Host, commandUri.EscapedPath())
			req := httplib.Post(notifyUrlStr)
			if len(params) > 0 {
				for k, m := range params {
					if len(m) > 0 {
						for _, v := range m {
							req.Param(k, v)
						}
					}
				}
			}
			resp, err := req.DoRequest()
			if err != nil {
				return "", "", err, false
			}
			defer resp.Body.Close()
			respByteArr, _ := ioutil.ReadAll(resp.Body)
			isTimeout := false
			if resp.StatusCode == 408 {
				isTimeout = true
			}
			return string(respByteArr), resp.Status, err, isTimeout
		}
	}
	return job
}

func (j *Job) Status() int {
	return j.status
}

func (j *Job) GetName() string {
	return j.name
}

func (j *Job) GetId() int {
	return j.id
}

func (j *Job) GetLogId() int64 {
	return j.logId
}

func (j *Job) Run() {
	if !j.Concurrent && j.status > 0 {
		beego.Warn(fmt.Sprintf("任务[%d]上一次执行尚未结束，本次被忽略。", j.id))
		return
	}

	defer func() {
		if err := recover(); err != nil {
			beegoLog.Error("errRecover", err, "\n", string(debug.Stack()))
		}
	}()

	if workPool != nil {
		workPool <- true
		defer func() {
			<-workPool
		}()
	}

	beego.Debug(fmt.Sprintf("开始执行任务: %d", j.id))

	j.status++
	defer func() {
		j.status--
	}()

	t := time.Now()
	timeout := time.Duration(time.Hour * 24)
	if j.task.Timeout > 0 {
		timeout = time.Second * time.Duration(j.task.Timeout)
	}

	cmdOut, cmdErr, err, isTimeout := j.runFunc(timeout)

	ut := time.Now().Sub(t) / time.Millisecond

	// 插入日志
	log := new(models.TaskLog)
	log.TaskId = j.id
	log.Output = cmdOut
	log.Error = cmdErr
	log.ProcessTime = int(ut)
	log.CreateTime = t.Unix()

	if isTimeout {
		log.Status = models.TASK_TIMEOUT
		log.Error = fmt.Sprintf("任务执行超过 %d 秒\n----------------------\n%s\n", int(timeout/time.Second), cmdErr)
	} else if err != nil {
		log.Status = models.TASK_ERROR
		log.Error = err.Error() + ":" + cmdErr
	}
	j.logId, _ = models.TaskLogAdd(log)

	// 更新上次执行时间
	j.task.PrevTime = t.Unix()
	j.task.ExecuteTimes++
	j.task.Update("PrevTime", "ExecuteTimes")

	// 发送邮件通知
	if (j.task.Notify == 1 && err != nil) || j.task.Notify == 2 {
		user, uerr := models.UserGetById(j.task.UserId)
		if uerr != nil {
			return
		}

		var (
			title         string
			receiverEmail string
			receiverName  string
		)

		data := make(map[string]interface{})
		data["task_id"] = j.task.Id
		data["username"] = user.UserName
		data["task_name"] = j.task.TaskName
		data["start_time"] = beego.Date(t, "Y-m-d H:i:s")
		data["process_time"] = float64(ut) / 1000
		data["output"] = cmdOut

		if isTimeout {
			title = fmt.Sprintf("任务执行结果通知 #%d: %s", j.task.Id, "超时")
			data["status"] = fmt.Sprintf("超时（%d秒）", int(timeout/time.Second))
		} else if err != nil {
			title = fmt.Sprintf("任务执行结果通知 #%d: %s", j.task.Id, "失败")
			data["status"] = "失败（" + err.Error() + "）"
		} else {
			title = fmt.Sprintf("任务执行结果通知 #%d: %s", j.task.Id, "成功")
			data["status"] = "成功"
		}

		//content := new(bytes.Buffer)
		//mailTpl.Execute(content, data)
		//ccList := make([]string, 0)
		//if j.task.NotifyEmail != "" {
		//	ccList = strings.Split(j.task.NotifyEmail, "\n")
		//}
		//if !mail.SendMail(user.Email, user.UserName, title, content.String(), ccList) {
		//	beegoLog.Error("发送邮件超时：", user.Email)
		//}
		ccList := make([]string, 0)
		if j.task.NotifyEmail != "" {
			ccList = strings.Split(j.task.NotifyEmail, "\n")
			receiverEmail = ccList[0]
			if len(ccList) > 0 {
				ccList = ccList[1:]
			}
			receiverEmailArr := strings.Split(ccList[0], "@")
			receiverName = receiverEmailArr[0]
			data["username"] = receiverName
		}
		data["brief"] = j.task.Description
		data["attach"] = ""
		content := new(bytes.Buffer)
		if libs.IsPathExist(j.task.NotifyEmailAttach) && libs.GetFilesize(j.task.NotifyEmailAttach) > mail.EMailAttachSize {
			attachFileName := libs.GetFilename(j.task.NotifyEmailAttach)
			saveKey := fmt.Sprintf("upload/webcron/%s", attachFileName)
			achivePwd := libs.GetTarOrZipPassword(j.task.Command)
			privateAccess := true
			if len(achivePwd) > 0 {
				privateAccess = false
			}
			accessUrl, errUpload := libs.SaveToQiniu(j.task.NotifyEmailAttach, saveKey, privateAccess)
			if errUpload == nil {
				data["attach"] = accessUrl
				if (len(achivePwd)) > 0 {
					data["pwd"] = achivePwd
					mailAttachWihtPwdTpl.Execute(content, data)
				} else {
					mailAttachTpl.Execute(content, data)
				}
			} else {
				mailTpl.Execute(content, data)
			}
			if !mail.SendMail(receiverEmail, receiverName, title, content.String(), ccList) {
				beegoLog.Error("发送邮件超时：", user.Email)
			}
		} else {
			mailTpl.Execute(content, data)
			if !mail.SendMailWithAttach(title, content.String(), j.task.NotifyEmailAttach, ccList) {
				beegoLog.Error("发送邮件超时：", user.Email)
			}
		}
		if libs.IsPathExist(j.task.NotifyEmailAttach) {
			os.Remove(j.task.NotifyEmailAttach)
		}
	}
	if j.task.TotalTimes > 0 && j.task.ExecuteTimes >= j.task.TotalTimes {
		j.task.Status = models.TaskStatus_END
		j.task.Update()
		RemoveJob(j.id)
	}
}

func isPathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		// glog.Info(err)
		return false
	}
	return true
}

func getFilesize(path string) int64 {
	fileinfo, err := os.Stat(path)
	if err == nil {
		return fileinfo.Size()
	}
	return 0
}
