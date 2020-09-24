/*
 * @Descripttion: uuid生成工具类
 * @Author: chenjun
 * @Date: 2020-07-29 16:29:13
 */

package utils

import (
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

/*
Get36UUID  创建36位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get36UUID() string {
	// 创建  error handling
	uidInfo := uuid.NewV4()
	return uidInfo.String()
}

/*
Get32UUID  创建32位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get32UUID() string {
	uid := Get36UUID()
	udiStr := strings.Replace(uid, "-", "", -1)
	return udiStr
}

/*
Get42UUID  创建42位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get42UUID() string {
	//得到当前时间戳
	timeUnix := time.Now().Unix()
	uidStr := strconv.FormatInt(timeUnix, 10) + Get32UUID()
	return uidStr
}

/*
Get45UUID  创建45位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get45UUID() string {
	//得到当前时间戳
	timeUnixNano := time.Now().UnixNano()
	timestamp := timeUnixNano / 1000000
	uidStr := strconv.FormatInt(timestamp, 10) + Get32UUID()
	return uidStr
}

/*
Get46UUID  创建46位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get46UUID() string {
	//得到当前时间
	now := time.Now()
	timeStr := now.Format("20060102150405")
	uidStr := timeStr + Get32UUID()
	return uidStr
}

/*
Get49UUID  创建49位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get49UUID() string {
	//得到当前时间
	now := time.Now()
	timeStr := now.Format("20060102150405.999")
	timeStr = strings.Replace(timeStr, ".", "", -1)
	uidStr := timeStr + Get32UUID()
	return uidStr
}

/*
Get51UUID  创建51位的uuid
 * @param {type}
 * @return: 返回字符串
*/
func Get51UUID() string {
	//得到当前时间
	now := time.Now()
	timeStr := now.Format("20060102150405.99999")
	timeStr = strings.Replace(timeStr, ".", "", -1)
	uidStr := timeStr + Get32UUID()
	return uidStr
}
