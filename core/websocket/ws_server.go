/*
 * @Descripttion: websocket服务端
 * @Author: chenjun
 * @Date: 2020-08-04 14:06:11
 */

package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"go-cmd-transfer/global"
	"go-cmd-transfer/utils"

	"github.com/gorilla/websocket"
	logger "github.com/sirupsen/logrus"
)

//WebsocketConnAll ws的所有连接 用于广播
var WebsocketConnAll map[string]*WsConnection

var (
	upgrader = websocket.Upgrader{
		// 读取存储空间大小
		ReadBufferSize: 4096,
		// 写入存储空间大小
		WriteBufferSize: 1024,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			/*if r.Method != "GET" {
				logger.Warn("request method is not GET")
				return false
			}
			if r.URL.Path != "/ws" {
				logger.Warn("websocket connect path error")
				return false
			}*/
			return true
		},
	}
)

func wsHandler(resp http.ResponseWriter, req *http.Request) {
	var (
		wsConn *websocket.Conn
		conn   *WsConnection
		msg    *Message
		err    error
	)
	// 完成ws协议的握手操作 完成http应答,在httpheader中放下如下参数 Upgrade:websocket 客户端告知升级连接为websocket
	wsConn, err = upgrader.Upgrade(resp, req, nil)
	if err != nil {
		logger.Error("升级为websocket失败", err.Error())
		// 获取连接失败直接返回
		return
	}
	connAddr := wsConn.RemoteAddr().String()
	logger.Infof("websocket客户端连接地址:%s", connAddr)
	connID := utils.Get49UUID()
	conn, err = InitConnection(wsConn, connID, connAddr)
	if err != nil {
		logger.Error("初始化websocket失败", err.Error())
		// 关闭当前连接
		conn.Close()
		return
	}
	// TODO 如果要控制连接数可以计算，wsConnAll长度
	// 存储连接信息，连接数保持一定数量，超过的部分不提供服务
	if conn != nil {
		WebsocketConnAll[connID] = conn
	}
	logger.Infof("websocket当前在线连接数:%d", len(WebsocketConnAll))

	// 启动线程，不断发消息
	go func() {
		var (
			err error
		)
		for {
			// 每隔一秒发送一次心跳  使用应用层心跳机制， 即客户端定时发送ping， 服务端响应pong。
			if err = conn.WriteMessage(websocket.PongMessage, []byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		for {
			if msg, err = conn.ReadMessage(); err != nil {
				logger.Error("读取websocket消息失败", err.Error())
				// 关闭当前连接
				conn.Close()
				return
			}
			if json.Valid(msg.data) == false {
				logger.Warn("读取websocket消息时，该消息不是一个json字符串，不做处理")
			} else {
				busData := global.BusinessData{}
				if err := json.Unmarshal(msg.data, &busData); err != nil {
					logger.Error("读取websocket消息时，该消息是一个json字符串，进行解析格式化，解析错误", err.Error())
					continue
				}
				logger.Info(busData.Protocol)
				if busData.Protocol == "socket" {
					global.SocketBusDataAllInfo[busData.UserID] = busData
				} else if busData.Protocol == "websocket" {
					global.WebSocketBusDataAllInfo[busData.UserID] = busData
				}
			}
		}
	}()

	//启动协程循环写入
	go func() {
		for {
			if len(global.WebSocketBusDataAllInfo) > 0 {
				tempData, err := json.Marshal(global.WebSocketBusDataAllInfo)
				if err != nil {
					logger.Error("发送socket消息时，将待转发的消息转换为json字符串错误", err.Error())
					continue
				}
				//发送给所有在线的客户端
				for _, client := range WebsocketConnAll {
					if err = client.WriteMessage(msg.messageType, tempData); err != nil {
						logger.Error("发送socket消息失败", err.Error())
						// 关闭当前连接
						conn.Close()
						return
					}
				}
				global.WebSocketBusDataAllInfo = make(map[string]global.BusinessData)
			}

		}
	}()

	/*for {
		if msg, err = conn.ReadMessage(); err != nil {
			logger.Error("读取websocket消息失败", err.Error())
			conn.Close()
			return
		}
		if err = conn.WriteMessage(msg.messageType, msg.data); err != nil {
			logger.Error("发送websocket消息失败", err.Error())
			conn.Close()
			return
		}
	}*/
}

//StartWebsocket 启动程序
func StartWebsocket(addrPort string) {
	WebsocketConnAll = make(map[string]*WsConnection)
	logger.Info("开启 WebSocket Server ...")
	// 当有请求访问ws时，执行此回调方法
	http.HandleFunc("/ws", wsHandler)
	// 监听127.0.0.1:端口
	uri := "0.0.0.0:" + addrPort
	err := http.ListenAndServe(uri, nil)
	if err != nil {
		logger.Error("监听并启动websocket失败", err.Error())
	}
}
