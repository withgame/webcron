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
	root := cobra.Command{Use: "webcron cmd, include server and version"}
	root.AddCommand(
		server.Cmd,
		version.Cmd,
	)
	root.Execute()
}
