/*
 * @Descripttion: socket服务端
 * @Author: chenjun
 * @Date: 2020-08-06 15:48:14
 */

package socket

import (
	"go-cmd-transfer/global"
	"go-cmd-transfer/utils"
	"net"

	jsoniter "github.com/json-iterator/go"
	logger "github.com/sirupsen/logrus"
)

//SocketConnAll 保存在线用户 cliAddr ===> Connection
var SocketConnAll map[string]*SConnection

//实例化工具类
var json = jsoniter.ConfigCompatibleWithStandardLibrary

//connHandler 处理用户连接
func serverConnHandler(conn net.Conn) {
	//conn是否有效
	if conn == nil {
		logger.Error("无效的 socket 连接")
		return
	}

	var (
		socketConn *SConnection
		data       []byte
		err        error
	)

	//连接标识
	connID := utils.Get49UUID()
	// 获取客户端的网络地址
	cliAddr := conn.RemoteAddr().String()
	socketConn, err = InitConnection(conn, connID, cliAddr)
	if err != nil {
		logger.Error("初始化socket失败", err.Error())
		// 关闭当前连接
		socketConn.Close()
		return
	}
	// 存储连接信息，连接数保持一定数量，超过的部分不提供服务
	if socketConn != nil {
		SocketConnAll[connID] = socketConn
	}
	logger.Infof("socket当前在线连接数:%d", len(SocketConnAll))

	go func() {
		for {
			if data, err = socketConn.ReadMessage(); err != nil {
				logger.Error("读取socket消息失败", err.Error())
				// 关闭当前连接
				socketConn.Close()
				return
			}
			if json.Valid(data) == false {
				logger.Warn("读取socket消息时，该消息不是一个json字符串，不做处理")
			} else {
				busData := global.BusinessData{}
				if err := json.Unmarshal(data, &busData); err != nil {
					logger.Error("读取socket消息时，该消息是一个json字符串，进行解析格式化，解析错误", err.Error())
					continue
				}
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
			if len(global.SocketBusDataAllInfo) > 0 {
				tempData, err := json.Marshal(global.SocketBusDataAllInfo)
				if err != nil {
					logger.Error("发送socket消息时，将待转发的消息转换为json字符串错误", err.Error())
					continue
				}
				//发送给所有在线的客户端
				for _, client := range SocketConnAll {
					if err = client.WriteMessage(tempData); err != nil {
						logger.Error("发送socket消息失败", err.Error())
						// 关闭当前连接
						socketConn.Close()
						return
					}
				}
				global.SocketBusDataAllInfo = make(map[string]global.BusinessData)
			}

		}
	}()
}

//ServerSocket 开启服务
func ServerSocket(addrPort string) {
	SocketConnAll = make(map[string]*SConnection)
	logger.Info("正在开启 Socket Server ...")
	// 监听127.0.0.1:端口
	uri := "0.0.0.0:" + addrPort
	listener, err := net.Listen("tcp", uri)
	if err != nil {
		logger.Warn("启动Socket服务出错", err.Error())
		return
	}

	logger.Info("开启 Socket Server成功")

	defer listener.Close()

	// 主协程，循环阻塞等待用户连接  ,接收多个用户的请求
	for {
		//接收来自 client 的连接,会阻塞
		conn, err := listener.Accept()

		if err != nil {
			logger.Warn("连接Socket出错", err.Error())
			// 关闭当前用户链接
			//conn.Close()
			continue
		}

		//处理用户连接 并发模式 新建一个协程,接收来自客户端的连接请求，一个连接 建立一个 conn，服务器资源有可能耗尽 BIO模式
		go serverConnHandler(conn)
	}

}
