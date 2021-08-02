package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const filePath = "./config.json"

type Config struct {
	HomeFolder     string
	AdaptersFolder string
	AppDataFolder  string
	DatabaseFile   string
	ReportFolder   string
}

func LoadConfiguration() Config {
	var config Config
	configFile, err := os.Open(filePath)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
