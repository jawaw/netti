package configs

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	logger "netti/pkg/log"
	"os"
	"strings"
	"time"
)

var (
	vip *viper.Viper = nil
	log *logger.Logger
)

// init .
func init() {
	var envStr = ""
	env := viper.New()
	log = logger.NewLogger()
	if env != nil {
		env.SetEnvPrefix("netti")
		env.AllowEmptyEnv(true)
		env.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		env.AutomaticEnv()

		envStr = env.GetString("env")
	}
	log.Infof("env: %s", envStr)

	vip = viper.New()
	if vip != nil {
		vip.SetConfigName("configure") //把json文件换成yaml文件，只需要配置文件名 (不带后缀)即可
		vip.AddConfigPath("./configs") //添加配置文件所在的路径
		vip.SetConfigType("yaml")      //设置配置文件类型
		err := vip.ReadInConfig()
		if err != nil {
			log.Fatalf("config file error: %s", err)
			os.Exit(-1)
		}

		vip.WatchConfig() //监听配置变化
		vip.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("配置发生变更：", e.Name)
		})
	}
}

// GetString .
func GetString(key string) string {
	if vip == nil {
		log.Fatalf("wrong env value")
		return ""
	}
	return vip.GetString(key)
}

// GetInteger .
func GetInteger(key string) int {
	if vip == nil {
		log.Fatalf("wrong env value")
		return 0
	}
	return vip.GetInt(key)
}

// GetDuration .
func GetDuration(key string) time.Duration {
	if vip == nil {
		log.Fatalf("wrong env value")
		return 0
	}
	return vip.GetDuration(key)
}
