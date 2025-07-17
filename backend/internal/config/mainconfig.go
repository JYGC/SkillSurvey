package config

import (
	"flag"

	"github.com/JYGC/SkillSurvey/internal/environment"
)

const (
	configPath     = "./config.json"
	testConfigPath = "./config.test.json"
)

type MainConfig struct {
	ConfigBase
	ServerPort string
}

func LoadMainConfig() MainConfig {
	var mainConfig MainConfig
	if flag.Lookup("test.v") == nil {
		JsonToConfig(&mainConfig, environment.AttachToExecutableDir(configPath))
	} else {
		JsonToConfig(&mainConfig, environment.AttachToExecutableDir(testConfigPath))
	}
	return mainConfig
}
