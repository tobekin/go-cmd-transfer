/*
 * @Descripttion: 配置文件信息
 * @Author: chenjun
 * @Date: 2020-07-30 15:43:55
 */

package config

//Server  服务配置
type Server struct {
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis"`
	System System `mapstructure:"system" json:"system" yaml:"system"`
	Log    Log    `mapstructure:"log" json:"log" yaml:"log"`
}

//System 信息
type System struct {
	Env           string `mapstructure:"env" json:"env" yaml:"env"`
	SocketPort    int    `mapstructure:"socket-port" json:"socketPport" yaml:"socket-port"`
	WebsocketPort int    `mapstructure:"websocket-port" json:"websocketPport" yaml:"websocket-port"`
}

//Redis 信息
type Redis struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	Database int    `mapstructure:"database" json:"database" yaml:"database"`
	Timeout  int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

//Log 信息
type Log struct {
	LogPath string `mapstructure:"log-path" json:"logPath" yaml:"log-path"`
	LogFile string `mapstructure:"log-file" json:"logFile" yaml:"log-file"`
	Level   string `mapstructure:"level" json:"level" yaml:"level"`
}
