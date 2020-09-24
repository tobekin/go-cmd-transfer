/*
 * @Descripttion: websocket配置
 * @Author: chenjun
 * @Date: 2020-08-04 10:26:49
 */

package websocket

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	maxMessageSize = 512
)

//Message 读写消息
type Message struct {
	// websocket.TextMessage 消息类型
	messageType int
	data        []byte
}

//WsConnection 连接信息
type WsConnection struct {
	// 存放websocket连接
	wsConn *websocket.Conn
	// 用于存放数据 读队列
	inChan chan *Message
	// 用于读取数据 写队列
	outChan chan *Message
	// 用于关闭连接
	closeChan chan byte
	// 对closeChan关闭上锁 避免重复关闭管道,加锁处理  互斥锁
	mutex sync.Mutex
	// chan是否被关闭 防止closeChan被关闭多次
	isClosed bool
	//连接标识
	wsID string
	// 网络地址
	addr string
}

//InitConnection 初始化长连接
func InitConnection(wsConn *websocket.Conn, connID string, connAddr string) (conn *WsConnection, err error) {
	conn = &WsConnection{
		wsConn:    wsConn,
		inChan:    make(chan *Message, 4096),
		outChan:   make(chan *Message, 4096),
		closeChan: make(chan byte, 1),
		isClosed:  false,
		wsID:      connID,
		addr:      connAddr,
	}

	// 处理器,发送定时信息，避免意外关闭
	go conn.processLoop()
	// 读协程
	go conn.readLoop()
	// 写协程
	go conn.writeLoop()
	return
}

//ReadMessage 读取消息队列中的消息
func (conn *WsConnection) ReadMessage() (msg *Message, err error) {
	logger.Infof("websocket读取消息，连接标识：%s，连接地址：%s", conn.wsID, conn.addr)
	//select是Go中的一个控制结构，类似于用于通信的switch语句。
	//每个case必须是一个通信操作，要么是发送要么是接收。
	//select随机执行一个可运行的case。如果没有case可运行，它将阻塞，直到有case可运行。一个默认的子句应该总是可运行的。
	select {
	// 从Channel中接收数据，并将数据赋值给msg
	case msg = <-conn.inChan:
		logger.Infof("websocket读取消息时，连接标识：%s，连接地址：%s，数据信息(消息类型为：%d,消息数据为：%s)", conn.wsID, conn.addr, msg.messageType, string(msg.data))
	case <-conn.closeChan:
		err = errors.New("connection is closed")
		logger.Errorf("websocket读取消息时，连接标识：%s，连接地址：%s，连接被关闭，错误信息：%s", conn.wsID, conn.addr, err.Error())
	}
	//如果return后面没有指定返回值，就用赋给“返回值变量”的值
	return
}

//WriteMessage 发送消息到队列中
func (conn *WsConnection) WriteMessage(messageType int, data []byte) (err error) {
	logger.Infof("websocket发送消息，连接标识：%s，连接地址：%s", conn.wsID, conn.addr)
	msg := &Message{messageType, data}
	select {
	// 发送值data到Channel中
	case conn.outChan <- msg:
		logger.Infof("websocket发送消息时，连接标识：%s，连接地址：%s，数据信息(消息类型为：%d,消息数据为：%s)", conn.wsID, conn.addr, msg.messageType, string(msg.data))
	case <-conn.closeChan:
		err = errors.New("connection is closed")
		logger.Errorf("websocket发送消息时，连接标识：%s，连接地址：%s，连接被关闭，错误信息：%s", conn.wsID, conn.addr, err.Error())
	}
	//当return后面为空是，函数声明时的 (err error) 会把 err 作为返回值，当 return 不为空时，会把 return 后面的值作为返回值
	return
}

//Close 关闭连接
func (conn *WsConnection) Close() {
	logger.Infof("websocket关闭连接，连接标识：%s，连接地址：%s", conn.wsID, conn.addr)
	// 线程安全的Close，可以并发多次调用也叫做可重入的Close
	conn.wsConn.Close()
	// 利用标记，让closeChan只关闭一次
	conn.mutex.Lock()
	logger.Infof("websocket关闭连接，连接标识：%s，连接地址：%s，当前连接是否关闭状态为：%t", conn.wsID, conn.addr, conn.isClosed)
	if conn.isClosed == false {
		// 关闭chan,但是chan只能关闭一次
		close(conn.closeChan)
		// 删除这个连接的变量
		delete(WebsocketConnAll, conn.wsID)
		conn.isClosed = true
	}
	//释放锁
	conn.mutex.Unlock()
}

// 处理器,发送定时信息，避免意外关闭  内部实现
func (conn *WsConnection) processLoop() {
	// 获取到消息队列中的消息，处理完成后，发送消息给客户端
	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			logger.Error("websocket获取消息出现错误", err.Error())
			break
		}
		logger.Info("websocket接收到消息", string(msg.data))
		// 修改以下内容把客户端传递的消息传递给处理程序
		err = conn.WriteMessage(msg.messageType, msg.data)
		if err != nil {
			logger.Error("websocket发送消息给客户端出现错误", err.Error())
			break
		}
	}
}

//读取消息队列中的消息 内部实现
func (conn *WsConnection) readLoop() {
	// 设置消息的最大长度
	conn.wsConn.SetReadLimit(maxMessageSize)
	conn.wsConn.SetReadDeadline(time.Now().Add(pongWait))
	for {
		// 读一个message
		msgType, data, err := conn.wsConn.ReadMessage()
		if err != nil {
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			logger.Errorf("websocket消息读取出现错误，连接标识：%s，连接地址：%s，错误信息为：%s", conn.wsID, conn.addr, err.Error())
			goto ERR
		}
		req := &Message{
			msgType,
			data,
		}
		// 放入请求队列,消息入栈 容易阻塞到这里，等待inChan有空闲的位置
		select {
		case conn.inChan <- req:
		case <-conn.closeChan:
			// closeChan关闭的时候执行
			goto ERR
		}
	}
ERR:
	conn.Close()
}

//发送消息队列中的消息 内部实现
func (conn *WsConnection) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	conn.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
	for {
		select {
		// 取一个应答
		case msg := <-conn.outChan:
			err := conn.wsConn.WriteMessage(msg.messageType, msg.data)
			if err != nil {
				logger.Errorf("websocket消息写入出现错误，连接标识：%s，连接地址：%s，错误信息为：%s", conn.wsID, conn.addr, err.Error())
				// 切断服务
				goto ERR
			}
		case <-conn.closeChan:
			// 获取到关闭通知
			goto ERR
		case <-ticker.C:
			// 出现超时情况
			conn.wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			logger.Errorf("websocket消息写入出现超时情况，连接标识：%s，连接地址：%s", conn.wsID, conn.addr)
			if err := conn.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorf("websocket消息写入出现超时情况，连接标识：%s，连接地址：%s，错误信息为：%s", conn.wsID, conn.addr, err.Error())
				goto ERR
			}
		}
	}
ERR:
	conn.Close()
}
