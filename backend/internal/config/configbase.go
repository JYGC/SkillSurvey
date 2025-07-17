package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type ConfigBase struct{}

func JsonToConfig(config any, fullFilePath string) {
	fileContents, err := os.ReadFile(fullFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	json.Unmarshal(fileContents, &config)
}
