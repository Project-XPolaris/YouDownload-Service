package config

import "github.com/spf13/viper"

var Instance Config

type Config struct {
	Addr        string
	DownloadDir string
	AuthEnable  bool
	PathEnable  bool
	YouPlusUrl  string
}

func ReadConfig() error {
	configer := viper.New()
	configer.AddConfigPath("./")
	configer.AddConfigPath("../")
	configer.SetConfigType("yaml")
	configer.SetConfigName("config")
	err := configer.ReadInConfig()
	if err != nil {
		return err
	}
	configer.SetDefault("addr", ":5700")
	configer.SetDefault("paths.download", "./dowmload")
	configer.SetDefault("youplus.auth", false)
	configer.SetDefault("youplus.pathEnable", false)
	configer.SetDefault("youplus.url", "http://localhost:8999")

	Instance = Config{
		Addr:        configer.GetString("addr"),
		DownloadDir: configer.GetString("paths.download"),
		AuthEnable:  configer.GetBool("youplus.auth"),
		PathEnable:  configer.GetBool("youplus.pathEnable"),
		YouPlusUrl:  configer.GetString("youplus.url"),
	}
	return nil
}
