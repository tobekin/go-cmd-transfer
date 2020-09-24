/*
 * @Descripttion: socket客户端
 * @Author: chenjun
 * @Date: 2020-08-06 17:24:34
 */

package socket

import (
	"bufio"
	"net"
	"os"

	logger "github.com/sirupsen/logrus"
)

//clientConnHandler 处理用户连接
func clientConnHandler(conn net.Conn, message string) {
	//1.conn是否有效
	if conn == nil {
		logger.Error("无效的 socket 连接")
		return
	}

	//切片缓冲 缓存 conn 中的数据
	buf := make([]byte, 4096)
	var breakfor int

	//服务器重连后，自动重新发送上次消息
	if message != "" {
		logger.Info("客户端自动重新建立socket连接")
		messageHandle(conn, message, buf)
	}

	//返回一个拥有 默认size 的reader，接收客户端输入
	reader := bufio.NewReader(os.Stdin)

	for {
		logger.Info("请输入客户端请求数据...")

		//客户端输入
		input, _ := reader.ReadString('\n')
		//发送消息
		breakfor = messageHandle(conn, input, buf)
		if breakfor == 1 {
			message = input
			conn.Close() //关闭连接
			break        //打断循环
		}
	}
}

//消息处理
func messageHandle(conn net.Conn, message string, buf []byte) int {
	//客户端请求数据写入 conn，并传输
	cnt, err := conn.Write(append(append([]byte("cmdmgt"), IntToBytes(len(message))...), message...))
	if cnt == 0 || err != nil {
		logger.Error("客户端请求数据写入socket连接通道中错误", err.Error())
		return 1 //退出当前循环
	}
	//服务器端返回的数据写入空buf  接收服务器回复的数据
	cnt, err = conn.Read(buf)
	if cnt == 0 || err != nil {
		logger.Error("客户端接收服务器回复的socket数据失败", err.Error())
		return 1 //退出当前循环
	}

	//回显服务器端回传的信息
	logger.Info("服务器端回复socket数据", string(buf[0:cnt]))
	return 0
}

//ClientConnect 客户端连接
func ClientConnect(message string, addrPort string) {
	uri := "0.0.0.0:" + addrPort
	conn, err := net.Dial("tcp", uri)
	if err != nil {
		logger.Error("客户端建立socket连接失败", err.Error())
		return
	}
	//启用协程
	go clientConnHandler(conn, message)
}
