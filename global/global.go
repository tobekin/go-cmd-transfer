/*
 * @Descripttion: 全局配置信息
 * @Author: chenjun
 * @Date: 2020-07-30 15:42:40
 */

package global

import (
	"go-cmd-transfer/config"

	"github.com/spf13/viper"
)

var (
	//CmdConfig 全局服务配置
	CmdConfig config.Server
	//CmdVp 配置文件
	CmdVp *viper.Viper
)
