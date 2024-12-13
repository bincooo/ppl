package wire

import (
	"github.com/iocgo/sdk"

	// ------>>> 将需要托管的实例构造器所属的包导入 <<<-----
	/*
	 *	使用了 Cobra / Inject / Bean 注解的包内会生成一个 `func Injects(container *sdk.Container) error` 函数
	 *	在此文件下：由于会扫描下划线的包并调用Injects，所以没有使用ioc注解的包请不要使用下划线别名
	 */
	_ "ppl/cmd/cobra"
	_ "ppl/gin"

	// 类似一些启动器装载
	_ "github.com/iocgo/sdk/cobra/scan" /* cobra 自动装配 */
	_ "github.com/iocgo/sdk/scan"       /* 内置环境变量 */
)

// @Gen()
func Injects(container *sdk.Container) error {
	panic("auto implements")
}
