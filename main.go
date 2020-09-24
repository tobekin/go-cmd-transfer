/*
 * @Descripttion: 主函数
 * @Author: chenjun
 * @Date: 2020-07-29 16:03:18
 */

package main

import (
	"go-cmd-transfer/core"
	"go-cmd-transfer/core/socket"
	"go-cmd-transfer/core/websocket"
	"go-cmd-transfer/global"
	"strconv"
)

func main() {

	//初始化配置
	core.InitYml()
	c := global.CmdConfig.Log
	core.InitLog(c.LogPath, c.LogFile, c.Level)

	//logger.WithFields(logger.Fields{"animal": "walrus"}).Info("A walrus appears")

	info := global.CmdConfig.System
	//socket.ClientConnect("test", strconv.Itoa(info.SocketPort))
	//开启协程运行socket服务
	go func() {
		socket.ServerSocket(strconv.Itoa(info.SocketPort))
	}()
	//开启websocket服务
	websocket.StartWebsocket(strconv.Itoa(info.WebsocketPort))
}
