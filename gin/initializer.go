package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/iocgo/sdk"
	"github.com/iocgo/sdk/env"
	"github.com/iocgo/sdk/router"
	"net/http"
)

// @Inject(lazy="false", name="ginInitializer")
func Initialized(env *env.Environment) sdk.Initializer {
	return sdk.InitializedWrapper(0, func(container *sdk.Container) (err error) {
		sdk.ProvideTransient(container, sdk.NameOf[*gin.Engine](), func() (engine *gin.Engine, err error) {
			if !env.GetBool("server.debug") {
				gin.SetMode(gin.ReleaseMode)
			}

			engine = gin.Default()
			{
				engine.Use(gin.Recovery())
				engine.Use(cros)
			}
			beans := sdk.ListInvokeAs[router.Router](container)
			for _, route := range beans {
				route.Routers(engine)
			}

			return
		})
		return
	})
}

func cros(gtx *gin.Context) {
	method := gtx.Request.Method
	gtx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	gtx.Header("Access-Control-Allow-Origin", "*") // 设置允许访问所有域
	gtx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
	gtx.Header("Access-Control-Allow-Headers", "*")
	gtx.Header("Access-Control-Expose-Headers", "*")
	gtx.Header("Access-Control-Max-Age", "172800")
	gtx.Header("Access-Control-Allow-Credentials", "false")
	gtx.Set("content-type", "application/json")
	if method == "OPTIONS" {
		gtx.Status(http.StatusOK)
		return
	}
	gtx.Next()
}
