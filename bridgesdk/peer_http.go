package bridgesdk

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

/*
* http 服务
 */
func createHttpServer() {
	r := gin.Default()
	gin.SetMode(gin.DebugMode)
	r.POST("/device/query", getDeviceOnlineState)
}

//获取设备是否在线的http请求
func getDeviceOnlineState(ctx *gin.Context) {
	appkey := ctx.PostForm("appkey")
	jpushtoken := ctx.PostForm("did")
	devid := ctx.PostForm("deviceid")
	if appkey == "" || jpushtoken == "" || devid == "" {
		ctx.String(http.StatusOK, "{\"status\":\"fail\",\"msg\":\"key len not match\"}")
		return
	}

	noteqtoapp := false
	for _, app := range config.Apppushconfigs.Apps {
		if appkey == app.Name {
			noteqtoapp = true
			break
		}
	}
	if !noteqtoapp {
		ctx.String(http.StatusOK, "{\"status\":\"fail\",\"msg\":\"appkey error\"}")
		return
	}

	ch := make(chan string)
	timeout := make(chan bool, 1)
	go func() { // 超时设置
		time.Sleep(time.Duration(5) * time.Second) // 等待5秒钟
		timeout <- true
	}()
	go addinfo(appkey, jpushtoken, devid, ch)

	select { // 如果超时前读取到数据的话就进行登录处理，如果没有数据的话就超时断开连接
	case result := <-ch:
		ctx.String(http.StatusOK, result)
		return
		// break
	case <-timeout:
		ctx.String(http.StatusOK, "{\"status\":\"timeout\"}")
		return
		// break
	}
}
