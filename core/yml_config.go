/*
 * @Descripttion: 核心配置
 * @Author: chenjun
 * @Date: 2020-07-30 16:33:53
 */

package core

import (
	"fmt"
	"go-cmd-transfer/global"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const defaultConfigFile = "config.yml"

//InitYml 解析配置文件
func InitYml() {
	v := viper.New()
	//v.AddConfigPath("./")
	//v.SetConfigName("config")
	v.SetConfigFile(defaultConfigFile)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := v.Unmarshal(&global.CmdConfig); err != nil {
			fmt.Println(err)
		}
	})
	if err := v.Unmarshal(&global.CmdConfig); err != nil {
		fmt.Println(err)
	}
	global.CmdVp = v
}
