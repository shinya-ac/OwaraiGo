package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
	_ "gopkg.in/ini.v1"
)

//go get gopkg.in/ini.v1←コンフィグファイル

type ConfigList struct {
	ApiKey    string
	ApiSecret string
}

var Config ConfigList

func init() {
	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		log.Printf("fail to road file%v", err)
		os.Exit(1)
	}

	Config = ConfigList{
		ApiKey:    cfg.Section("Slack").Key("hoge").String(),
		ApiSecret: cfg.Section("Slack").Key("fuga").String(),
	}

}
