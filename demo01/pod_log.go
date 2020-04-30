package demo01
import (
	"arena-serve/k8s"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// log流式处理器
type streamHandlerLog struct {
	wsConn *WsConnection
	send   chan []byte
	Offset int64
}

// 向web端输出
func (handler *streamHandlerLog) Write(name string) {
	var (
		count int
		err   error
		logData *k8s.LodData
	)
	logData, err = k8s.GetLogs(name, -1, "") //获取job日志
	//clientset.CoreV1().Pods("default").GetLogs("nginx-deployment-5c689d88bb-sm9nl", &core_v1.PodLogOptions{Container: "nginx", TailLines: &tailLines}) //可获取pod日志
	//发送给websocket client
	log.Info(len(logData.Content))
	if len(logData.Content) > 1024*1024 {
		var TmpLogData *k8s.LodData
		TmpLogData, err = k8s.GetLogs(name, 1000, "") //获取job日志
		handler.wsConn.WsWrite(websocket.TextMessage, []byte("由于日志量太大，只显示部分开头和结尾\n"))
		handler.wsConn.WsWrite(websocket.TextMessage, []byte(fmt.Sprintf("查看详细日志,请登录pod: %s\n",logData.PodName[0])))
		handler.wsConn.WsWrite(websocket.TextMessage, []byte("日志文件路径: "))
		for _,v := range logData.ContainerName {
			cname := strings.Split(v,"docker://")[1]
			filePath := path.Join("/var/lib/docker/containers/",cname, cname+"-json.log")
			handler.wsConn.WsWrite(websocket.TextMessage, []byte(fmt.Sprintf("%s\n",filePath)))
		}
		handler.wsConn.WsWrite(websocket.TextMessage, []byte("-----------------\n"))
		logData.Content = append(logData.Content[:51200],TmpLogData.Content...)
	}
	if err != nil {
		handler.wsConn.WsWrite(websocket.TextMessage, []byte(err.Error()))
		time.Sleep(1 * time.Second)
		return
	}
	handler.wsConn.WsWrite(websocket.TextMessage, logData.Content)
	for {
		time.Sleep(time.Second * 5)
		log.Info(count)
		if !logData.Flag || count > 24 {   //   超过2min没有日志输出则主动断开 或者 任务已完成
			handler.wsConn.WsClose()
			break
		}
		logData, err = k8s.GetLogs(name, -1, strconv.Itoa(5))
		log.Info(len(logData.Content),logData.Flag)
		if err := handler.wsConn.WsWrite(websocket.TextMessage, logData.Content); err != nil {
			log.Info(err.Error())         //客户端主动断开
			break
		}
		if len(logData.Content) == 0 {
			count++
			continue
		}
		count = 0
	}
	return
}

func HandlerLog(resp http.ResponseWriter, req *http.Request) {
	var (
		wsConn  *WsConnection
		handler *streamHandlerLog
		err     error
	)
	// 解析GET参数
	if err = req.ParseForm(); err != nil {
		return
	}
	name := req.Form.Get("name")
	log.Info("Websocket logs jobName: ", name)

	// 得到websocket长连接
	if wsConn, err = InitWebsocket(resp, req); err != nil {
		return
	}
	handler = &streamHandlerLog{
		wsConn: wsConn,
	}
	handler.Write(name)
	return
}
