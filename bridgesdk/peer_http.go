package bridgesdk

import (
	"github.com/gin-gonic/gin"
	"gopeer/global"
	"net/http"
	"time"
)

/*
* http 服务
 */
func createHttpServer() {
	r := gin.Default()
	gin.SetMode(gin.DebugMode)
	r.POST("/device/state", getDeviceOnlineState)
}

//获取设备是否在线的http请求
func getDeviceOnlineState(ctx *gin.Context) {
	appkey := ctx.PostForm("appkey")
	did := ctx.PostForm("did")

	if appkey == "" || did == "" {
		ctx.String(http.StatusOK, "{\"status\":\"fail\",\"msg\":\"key len not match\"}")
		return
	}

	if appkey == global.HttpKey {
		ctx.String(http.StatusOK, "{\"status\":\"fail\",\"msg\":\"appkey error\"}")
		return
	}

	ch := make(chan string)
	timeout := make(chan bool, 1)
	go func() { // 超时设置
		time.Sleep(time.Duration(5) * time.Second) // 等待5秒钟
		timeout <- true
	}()

	//查询
	go func() {
		if wsBuffer == nil {
			ch <- "{\"status\":\"not support\"}"
		}
		if m == nil {
			ch <- "{\"status\":\"lock error\"}"
		}

		m.RLock()
		_, ok := wsBuffer[did]
		m.RUnlock()
		if ok {
			ch <- "{\"status\":\"online\"}"
		} else {
			ch <- "{\"status\":\"offline\"}"
		}
	}()

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
