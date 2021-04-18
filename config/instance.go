package config

import "github.com/spf13/viper"

var Instance Config

type Config struct {
	Addr        string
	DownloadDir string
	AuthEnable  bool
	AuthUrl     string
	AuthRel     string
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
	configer.SetDefault("auth.enable", false)
	configer.SetDefault("auth.url", "http://localhost:8999")
	configer.SetDefault("auth.rel", "http://localhost:8999/user/auth")

	Instance = Config{
		Addr:        configer.GetString("addr"),
		DownloadDir: configer.GetString("paths.download"),
		AuthEnable:  configer.GetBool("auth.enable"),
		AuthUrl:     configer.GetString("auth.url"),
		AuthRel:     configer.GetString("auth.rel"),
	}
	return nil
}
