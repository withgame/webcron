/*
* File:    main.go
* Created: 2020-04-30 16:34
* Authors: MS geek.snail@qq.com
* Copyright (c) 2013 - 2020 虾游网络科技有限公司 版权所有
 */

package main

import (
	_ "net/http/pprof" // 注册 pprof 接口

	"webcron/cmd/server"
	"webcron/cmd/version"

	"github.com/spf13/cobra"
	_ "go.uber.org/automaxprocs" // 根据容器配额设置 maxprocs
)

var (
	a string
	v string
	c string
	d string
)

func main() {
	version.BinAppName = a
	version.BinBuildCommit = c
	version.BinBuildVersion = v
	version.BinBuildDate = d
	root := cobra.Command{Use: "sniper"}
	root.AddCommand(
		server.Cmd,
		version.Cmd,
	)
	root.Execute()
}
