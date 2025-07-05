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
	ServerPort   string
	IsProduction bool
	ResultUiRoot string
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
