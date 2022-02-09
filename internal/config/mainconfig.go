package config

import (
	"flag"
)

const (
	configPath     = "./config.json"
	testConfigPath = "./config.test.json"
)

type MainConfig struct {
	ConfigBase
	HomeFolder     string
	AdaptersFolder string
	AppDataFolder  string
	DatabaseFile   string
	ReportFolder   string
	ServerPort     string
	ErrorLogFile   string
}

func LoadMainConfig() MainConfig {
	var mainConfig MainConfig
	if flag.Lookup("test.v") == nil {
		JsonToConfig(&mainConfig, configPath)
	} else {
		JsonToConfig(&mainConfig, testConfigPath)
	}
	return mainConfig
}
