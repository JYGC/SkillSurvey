package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type IConfig interface{}

type ConfigBase struct{}

func JsonToConfig(config IConfig, filePath string) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	json.Unmarshal(fileContents, &config)
}
