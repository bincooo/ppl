package cobra

import (
	"github.com/gin-gonic/gin"
	"github.com/iocgo/sdk"
	"github.com/iocgo/sdk/cobra"
	"github.com/iocgo/sdk/env"
	"github.com/iocgo/sdk/inited"
	"github.com/sirupsen/logrus"
	"ppl/core"
	"ppl/core/logger"
	"strconv"
)

type RootCommand struct {
	container *sdk.Container
	env       *env.Environment
	engine    *gin.Engine

	Port     int    `cobra:"port" short:"p" usage:"服务端口 port"`
	LogLevel string `cobra:"log" short:"L" usage:"日志级别: trace|debug|info|warn|error"`
	LogPath  string `cobra:"log-path" usage:"日志路径 log path"`
	Proxied  string `cobra:"proxies" short:"P" usage:"本地代理 proxies"`
}

// @Cobra(name="cobra"
//
//	version = "v1.0.0"
//	use	    = "ppl"
//	short   = "代理池"
//	long    = "项目地址: https://www.github.com/bincooo/ppl"
//	run     = "Run"
//
// )
func New(container *sdk.Container, engine *gin.Engine, config string) (rc cobra.ICobra, err error) {
	environ, err := sdk.InvokeBean[*env.Environment](container, "")
	if err != nil {
		return
	}
	rc = cobra.ICobraWrapper(&RootCommand{
		container: container,
		env:       environ,
		engine:    engine,

		Port:     8080,
		LogLevel: "error",
		LogPath:  "log",
	}, config)
	return
}

func (rc *RootCommand) Run(cmd *cobra.Command, args []string) {
	println("root command running ...  ")
	logger.InitLogger(rc.LogPath, LogLevel(rc.LogLevel))
	Initialized(rc)
	inited.Initialized(false, rc.env)
	go core.Run(rc.env)

	// gin
	if rc.env.GetBool("server.debug") {
		println(rc.container.HealthLogger())
	}

	addr := ":" + strconv.Itoa(rc.Port)
	println("Listening and serving HTTP on 0.0.0.0" + addr)
	if err := rc.engine.Run(addr); err != nil {
		panic(err)
	}
}

func Initialized(rc *RootCommand) {
	if rc.Port != 0 {
		rc.env.Set("server.port", rc.Port)
	}
	if rc.Proxied != "" {
		rc.env.Set("server.proxied", rc.Proxied)
	}
}

func LogLevel(lv string) logrus.Level {
	switch lv {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
