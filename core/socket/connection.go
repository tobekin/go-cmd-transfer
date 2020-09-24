/*
 * @Descripttion: socket配置
 * @Author: chenjun
 * @Date: 2020-08-13 11:40:13
 */

package socket

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"time"

	logger "github.com/sirupsen/logrus"
)

const (
	// 允许等待的写入时间
	writeWait = 1000 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 6000 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10240

	// 固定头部
	headerInfo = "cmdmgt"
	// 固定头部长度
	headerInfoLength = len(headerInfo)
	// 保存数据长度
	saveDataLength = 4
)

//SConnection 连接信息
type SConnection struct {
	// 存放socket连接
	socketConn net.Conn
	// 用于存放数据 读队列
	inChan chan []byte
	// 用于读取数据 写队列
	outChan chan []byte
	// 用于关闭连接
	closeChan chan byte
	// 对closeChan关闭上锁 避免重复关闭管道,加锁处理  互斥锁
	mutex sync.Mutex
	// chan是否被关闭 防止closeChan被关闭多次
	isClosed bool
	//连接标识
	sid string
	// 网络地址
	addr string
}

//InitConnection 初始化长连接
func InitConnection(sConn net.Conn, connID string, connAddr string) (conn *SConnection, err error) {
	conn = &SConnection{
		socketConn: sConn,
		inChan:     make(chan []byte, 4096),
		outChan:    make(chan []byte, 4096),
		closeChan:  make(chan byte, 1),
		isClosed:   false,
		sid:        connID,
		addr:       connAddr,
	}

	// 读协程
	go conn.readLoop()
	// 写协程
	go conn.writeLoop()
	return
}

//ReadMessage 读取消息队列中的消息
func (conn *SConnection) ReadMessage() (data []byte, err error) {
	logger.Infof("socket读取消息，连接标识：%s，连接地址：%s", conn.sid, conn.addr)
	//select是Go中的一个控制结构，类似于用于通信的switch语句。
	//每个case必须是一个通信操作，要么是发送要么是接收。
	//select随机执行一个可运行的case。如果没有case可运行，它将阻塞，直到有case可运行。一个默认的子句应该总是可运行的。
	select {
	// 从Channel中接收数据，并将数据赋值给msg
	case data = <-conn.inChan:
		logger.Infof("socket读取消息时，连接标识：%s，连接地址：%s，数据信息为：%s", conn.sid, conn.addr, string(data))
	case <-conn.closeChan:
		err = errors.New("connection is closed")
		logger.Errorf("socket读取消息时，连接标识：%s，连接地址：%s，连接被关闭，错误信息：%s", conn.sid, conn.addr, err.Error())
	}
	//如果return后面没有指定返回值，就用赋给“返回值变量”的值
	return
}

//WriteMessage 发送消息到队列中
func (conn *SConnection) WriteMessage(data []byte) (err error) {
	logger.Infof("socket发送消息，连接标识：%s，连接地址：%s", conn.sid, conn.addr)
	select {
	// 发送值data到Channel中
	case conn.outChan <- data:
		logger.Infof("socket发送消息时，连接标识：%s，连接地址：%s，数据信息为：%s", conn.sid, conn.addr, string(data))
	case <-conn.closeChan:
		err = errors.New("connection is closed")
		logger.Errorf("socket发送消息时，连接标识：%s，连接地址：%s，连接被关闭，错误信息：%s", conn.sid, conn.addr, err.Error())
	}
	//当return后面为空是，函数声明时的 (err error) 会把 err 作为返回值，当 return 不为空时，会把 return 后面的值作为返回值
	return
}

//Close 关闭连接
func (conn *SConnection) Close() {
	logger.Infof("socket关闭连接，连接标识：%s，连接地址：%s", conn.sid, conn.addr)
	// 线程安全的Close，可以并发多次调用也叫做可重入的Close
	conn.socketConn.Close()
	// 利用标记，让closeChan只关闭一次
	conn.mutex.Lock()
	logger.Infof("socket关闭连接，连接标识：%s，连接地址：%s，当前连接是否关闭状态为：%t", conn.sid, conn.addr, conn.isClosed)
	if conn.isClosed == false {
		// 关闭chan,但是chan只能关闭一次
		close(conn.closeChan)
		// 删除这个连接的变量
		delete(SocketConnAll, conn.sid)
		conn.isClosed = true
	}
	//释放锁
	conn.mutex.Unlock()
}

//读取消息队列中的消息 内部实现
func (conn *SConnection) readLoop() {
	//消息格式为  头部信息+数据长度（4）个字节+数据
	conn.socketConn.SetReadDeadline(time.Now().Add(pongWait))
	// 数据缓冲
	databuf := make([]byte, maxMessageSize)
	//循环读取网络数据流
	for {
		//网络数据流读入 buffer
		cnt, err := conn.socketConn.Read(databuf)
		logger.Infof("socket消息读取，连接标识：%s，连接地址：%s，一条消息的长度为：%d", conn.sid, conn.addr, cnt)
		//数据读尽、读取错误 socket连接错误
		if err != nil {
			logger.Errorf("socket消息读取出现错误，连接标识：%s，连接地址：%s，错误信息为：%s", conn.sid, conn.addr, err.Error())
			goto ERR
		}
		//解包
		unpackLoop(databuf[0:cnt], conn)
	}
ERR:
	conn.Close()
}

//发送消息队列中的消息 内部实现
func (conn *SConnection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	conn.socketConn.SetWriteDeadline(time.Now().Add(writeWait))
	for {
		select {
		// 取一个应答
		case data := <-conn.outChan:
			//封包
			data = packetLoop(data)
			_, err := conn.socketConn.Write(data)
			if err != nil {
				logger.Errorf("socket消息写入出现错误，连接标识：%s，连接地址：%s，错误信息为：%s", conn.sid, conn.addr, err.Error())
				// 切断服务
				goto ERR
			}
			//logger.Infof("socket 写入的消息为：%s", string(data[0:cnt]))
		case <-conn.closeChan:
			// 获取到关闭通知
			goto ERR
		case <-ticker.C:
			// 出现超时情况
			conn.socketConn.SetWriteDeadline(time.Now().Add(writeWait))
			logger.Errorf("socket消息写入出现超时情况，连接标识：%s，连接地址：%s", conn.sid, conn.addr)
			if _, err := conn.socketConn.Write([]byte("time out leave out")); err != nil {
				logger.Errorf("socket消息写入出现超时情况，连接标识：%s，连接地址：%s，错误信息为：%s", conn.sid, conn.addr, err.Error())
				goto ERR
			}
		}
	}
ERR:
	conn.Close()
}

//封包
func packetLoop(message []byte) []byte {
	return append(append([]byte(headerInfo), IntToBytes(len(message))...), message...)
}

//解包
func unpackLoop(buffer []byte, conn *SConnection) {
	length := len(buffer)
	// 检查超长消息
	if length > maxMessageSize {
		logger.Errorf("socket消息读取出现错误，连接标识：%s，连接地址：%s，消息长度太长：%d", conn.sid, conn.addr, length)
		return
	}
	//如果消息长度不够直接返回
	if length < headerInfoLength+saveDataLength {
		// 放入请求队列,消息入栈 容易阻塞到这里，等待inChan有空闲的位置
		select {
		case conn.inChan <- buffer:
		case <-conn.closeChan:
			// closeChan关闭的时候执行
			conn.Close()
		}
		return
	}

	var i int
	//消息分割循环
	var index int
	for i = 0; i < length; i = i + 1 {
		if length < i+headerInfoLength+saveDataLength {
			break
		}
		// 消息头
		if string(buffer[i:i+headerInfoLength]) == headerInfo {
			index++
			//头部信息+数据长度
			dataIndex := i + headerInfoLength + saveDataLength
			logger.Infof("socket消息解包读取时，连接标识：%s，连接地址：%s，一条消息的第%d个包的数据位置：%d", conn.sid, conn.addr, index, dataIndex)
			//消息长度
			messageLength := BytesToInt(buffer[i+headerInfoLength : dataIndex])
			logger.Infof("socket消息解包读取时，连接标识：%s，连接地址：%s，一条消息的第%d个包的数据长度：%d", conn.sid, conn.addr, index, messageLength)
			//提取数据
			if length < dataIndex+messageLength {
				logger.Warnf("socket消息解包读取时，连接标识：%s，连接地址：%s，一条消息的第%d个包的数据截止位置超长", conn.sid, conn.addr, index)
				index = 0
				break
			}
			data := buffer[dataIndex : dataIndex+messageLength]
			logger.Infof("socket消息解包读取时，连接标识：%s，连接地址：%s，一条消息的第%d个包的数据长度：%d，数据信息为：%s", conn.sid, conn.addr, index, messageLength, string(data))
			// 放入请求队列,消息入栈 容易阻塞到这里，等待inChan有空闲的位置
			select {
			case conn.inChan <- data:
			case <-conn.closeChan:
				// closeChan关闭的时候执行
				conn.Close()
			}
			//读取下一个消息
			i += headerInfoLength + saveDataLength + messageLength - 1
		}
	}
	//一个包都没有解析到
	if index == 0 {
		logger.Warnf("socket消息解包读取时，连接标识：%s，连接地址：%s，一条消息的一个包的数据都未能解析", conn.sid, conn.addr)
		select {
		case conn.inChan <- buffer:
		case <-conn.closeChan:
			// closeChan关闭的时候执行
			conn.Close()
		}
	}
}

//IntToBytes 整型转换成字节
func IntToBytes(length int) []byte {
	x := int32(length)

	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

//BytesToInt 字节转换成整型
func BytesToInt(data []byte) int {
	// 消息缓冲
	bytesBuffer := bytes.NewBuffer(data)

	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}
