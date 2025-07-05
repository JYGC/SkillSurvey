package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type IConfig interface{}

type ConfigBase struct{}

func attachToExecutableDir(fileName string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := filepath.Dir(exe)
	fullFilePath := filepath.Join(exeDir, fileName)
	return fullFilePath
}

func JsonToConfig(config IConfig, fileName string) {
	fullFilePath := attachToExecutableDir(fileName)
	fileContents, err := os.ReadFile(fullFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	json.Unmarshal(fileContents, &config)
}
