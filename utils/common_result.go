/*
 * @Descripttion: 返回结果工具类
 * @Author: chenjun
 * @Date: 2020-07-31 14:36:50
 */

package utils

import (
	"encoding/json"

	logger "github.com/sirupsen/logrus"
)

/*
CommonResultResp 返回结构体
 * @param {type}
 * @return:
*/
type CommonResultResp struct {
	Status  bool        `json:"status"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

/*
CommonResult json数据字符串
 * @param status  响应状态
 * @param code    响应编码
 * @param message 响应消息
 * @param data    响应数据
 * @return: json数据字符串
*/
func CommonResult(status bool, code string, message string, data interface{}) string {
	// 将结构体解析为字符串
	jsonTemp, err := json.Marshal(CommonResultResp{
		status,
		code,
		message,
		data,
	})

	if err != nil {
		logger.Error("响应数据转换json字符串错误", err)
		return ""
	}
	jsonStr := string(jsonTemp)
	logger.Infof("响应数据转换json字符串为：%s", jsonStr)
	return jsonStr
}

/*
SuccessWithData 返回成功数据
 * @param data 响应数据
 * @return: 成功数据
*/
func SuccessWithData(data interface{}) string {
	return CommonResult(true, "0000", "ok", data)
}

/*
SuccessWithMessage 返回成功消息
 * @param message 响应消息
 * @return: 成功消息
*/
func SuccessWithMessage(message string) string {
	return CommonResult(true, "0000", message, map[string]interface{}{})
}

/*
SuccessDataMessage 返回成功消息数据
 * @param message 响应消息
 * @param data 响应数据
 * @return: 成功消息数据
*/
func SuccessDataMessage(message string, data interface{}) string {
	return CommonResult(true, "0000", message, data)
}

/*
SuccessCodeDataMessage 返回成功消息数据
 * @param code 响应编码
 * @param message 响应消息
 * @param data 响应数据
 * @return: 成功消息数据
*/
func SuccessCodeDataMessage(code string, message string, data interface{}) string {
	return CommonResult(true, code, message, data)
}

/*
FailWithMessage 返回失败消息
 * @param message 响应消息
 * @return: 失败消息
*/
func FailWithMessage(message string) string {
	return CommonResult(false, "9999", message, map[string]interface{}{})
}

/*
FailCodeMessage 返回失败消息
 * @param code 响应编码
 * @param message 响应消息
 * @return: 失败消息
*/
func FailCodeMessage(code string, message string) string {
	return CommonResult(false, code, message, map[string]interface{}{})
}

/*
FailCodeDataMessage 返回失败消息
 * @param code 响应编码
 * @param message 响应消息
 * @param data 响应数据
 * @return: 失败消息
*/
func FailCodeDataMessage(code string, message string, data interface{}) string {
	return CommonResult(false, code, message, data)
}
