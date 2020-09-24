/*
 * @Descripttion: 全局业务数据配置
 * @Author: chenjun
 * @Date: 2020-08-18 16:41:39
 */

package global

//SocketBusDataAllInfo  socket业务数据集合
var SocketBusDataAllInfo = make(map[string]BusinessData)

//WebSocketBusDataAllInfo  websocket业务数据集合
var WebSocketBusDataAllInfo = make(map[string]BusinessData)

//BusinessData 业务数据报文
type BusinessData struct {
	Protocol string      `json:"protocol"` // 协议 socket/websocket
	SourceID string      `json:"sourceId"` // 接入端标识
	UserID   string      `json:"userId"`   // 用户账号
	OpType   string      `json:"opType"`   // 操作类型
	Data     interface{} `json:"data"`     // 数据
}
