package config

const MainConfigFilePath = "./config.json"

type MainConfig struct {
	ConfigBase
	HomeFolder     string
	AdaptersFolder string
	AppDataFolder  string
	DatabaseFile   string
	ReportFolder   string
}

func LoadMainConfig() MainConfig {
	var mainConfig MainConfig
	JsonToConfig(&mainConfig, MainConfigFilePath)
	return mainConfig
}
