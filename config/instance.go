package config

import "github.com/spf13/viper"

var Instance Config

type Config struct {
	Addr        string
	TmpDir      string
	DownloadDir string
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
	configer.SetDefault("paths.tmp", "./tmp")
	configer.SetDefault("paths.download", "./dowmload")

	Instance = Config{
		Addr:        configer.GetString("addr"),
		TmpDir:      configer.GetString("paths.tmp"),
		DownloadDir: configer.GetString("paths.download"),
	}
	return nil
}
