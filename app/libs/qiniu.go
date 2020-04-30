/*
* File:    qiniu.go
* Created: 2019-01-07 16:53
* Authors: MS geek.snail@qq.com
* Copyright (c) 2013 - 2019 青木文化传播有限公司 版权所有
 */

package libs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/qiniu/api.v7/v7/auth"
	"github.com/qiniu/api.v7/v7/storage"
)

func SaveToQiniu(target, key string, isPrivate bool) (accessUrl string, err error) {
	if len(webCronQnAK) == 0 {
		err = errors.New("missing qn ak")
		return
	}
	if len(webCronQnSK) == 0 {
		err = errors.New("missing qn sk")
		return
	}
	if len(webCronQnHub) == 0 {
		err = errors.New("missing qn hub")
		return
	}
	hub := webCronQnHub
	if isPrivate {
		hub = webCronQnPrivateHub
	}
	saveName := fmt.Sprintf("%s:%s", hub, key)
	putPolicy := storage.PutPolicy{
		Scope:   fmt.Sprintf("%s:%s", hub, key),
		SaveKey: saveName,
	}
	mac := auth.New(webCronQnAK, webCronQnSK)
	deadline := time.Now().Add(time.Second * 1800).Unix() //0.5小时有效期
	if isPrivate {
		accessUrl = storage.MakePrivateURL(mac, qnPrivateAccessDomain, key, deadline)
	} else {
		accessUrl = storage.MakePublicURL(qnAccessDomain, key)
	}
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	if len(webCronQnZone) == 0 {
		reg, b := storage.GetRegionByID(storage.RegionID(webCronQnZone))
		if !b {
			cfg.Zone = &storage.ZoneHuanan
		}
		cfg.Zone = &reg
	} else {
		cfg.Zone = &storage.ZoneHuanan
	}
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false
	resumeUploader := storage.NewResumeUploader(&cfg)
	ret := storage.PutRet{}
	//RETRY:
	err = resumeUploader.PutFile(context.Background(), &ret, upToken, key, target, nil)
	if err != nil {
		return
	}
	err = os.Remove(target)
	if err != nil {
		return
	}
	return
}
