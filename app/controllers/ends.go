package controllers

import (
	"strconv"
	"time"

	"webcron/app/jobs"
	"webcron/app/libs"
	"webcron/app/models"

	"github.com/astaxie/beego"
)

type EndsController struct {
	BaseController
}

// 任务列表
func (this *EndsController) List() {
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}
	groupId, _ := this.GetInt("groupid")
	filters := make([]interface{}, 0)
	filters = append(filters, "status", 2)
	if groupId > 0 {
		filters = append(filters, "group_id", groupId)
	}
	result, count := models.TaskGetList(page, this.pageSize, filters...)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["name"] = v.TaskName
		row["cron_spec"] = v.CronSpec
		row["status"] = v.Status
		row["description"] = v.Description

		e := jobs.GetEntryById(v.Id)
		if e != nil {
			row["next_time"] = beego.Date(e.Next, "Y-m-d H:i:s")
			row["prev_time"] = "-"
			if e.Prev.Unix() > 0 {
				row["prev_time"] = beego.Date(e.Prev, "Y-m-d H:i:s")
			} else if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			}
			row["running"] = 1
		} else {
			row["next_time"] = "-"
			if v.PrevTime > 0 {
				row["prev_time"] = beego.Date(time.Unix(v.PrevTime, 0), "Y-m-d H:i:s")
			} else {
				row["prev_time"] = "-"
			}
			row["running"] = 0
		}
		list[k] = row
	}

	// 分组列表
	groups, _ := models.TaskGroupGetList(1, 100)

	this.Data["pageTitle"] = "任务列表"
	this.Data["list"] = list
	this.Data["groups"] = groups
	this.Data["groupid"] = groupId
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("EndsController.List", "groupid", groupId), true).ToString()
	this.display()
}

// 任务执行日志列表
func (this *EndsController) Logs() {
	taskId, _ := this.GetInt("id")
	page, _ := this.GetInt("page")
	if page < 1 {
		page = 1
	}

	task, err := models.TaskGetById(taskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	result, count := models.TaskLogGetList(page, this.pageSize, "task_id", task.Id)

	list := make([]map[string]interface{}, len(result))
	for k, v := range result {
		row := make(map[string]interface{})
		row["id"] = v.Id
		row["start_time"] = beego.Date(time.Unix(v.CreateTime, 0), "Y-m-d H:i:s")
		row["process_time"] = float64(v.ProcessTime) / 1000
		row["ouput_size"] = libs.SizeFormat(float64(len(v.Output)))
		row["status"] = v.Status
		list[k] = row
	}

	this.Data["pageTitle"] = "任务执行日志"
	this.Data["list"] = list
	this.Data["task"] = task
	this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("EndsController.Logs", "id", taskId), true).ToString()
	this.display()
}

// 查看日志详情
func (this *EndsController) ViewLog() {
	id, _ := this.GetInt("id")

	taskLog, err := models.TaskLogGetById(id)
	if err != nil {
		this.showMsg(err.Error())
	}

	task, err := models.TaskGetById(taskLog.TaskId)
	if err != nil {
		this.showMsg(err.Error())
	}

	data := make(map[string]interface{})
	data["id"] = taskLog.Id
	data["output"] = taskLog.Output
	data["error"] = taskLog.Error
	data["start_time"] = beego.Date(time.Unix(taskLog.CreateTime, 0), "Y-m-d H:i:s")
	data["process_time"] = float64(taskLog.ProcessTime) / 1000
	data["ouput_size"] = libs.SizeFormat(float64(len(taskLog.Output)))
	data["status"] = taskLog.Status

	this.Data["task"] = task
	this.Data["data"] = data
	this.Data["pageTitle"] = "查看日志"
	this.display()
}

// 批量操作日志
func (this *EndsController) LogBatch() {
	action := this.GetString("action")
	ids := this.GetStrings("ids")
	if len(ids) < 1 {
		this.ajaxMsg("请选择要操作的项目", MSG_ERR)
	}
	for _, v := range ids {
		id, _ := strconv.Atoi(v)
		if id < 1 {
			continue
		}
		switch action {
		case "delete":
			models.TaskLogDelById(id)
		}
	}

	this.ajaxMsg("", MSG_OK)
}
