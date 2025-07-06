package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

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

func JsonToConfig(config any, fileName string) {
	fullFilePath := attachToExecutableDir(fileName)
	fileContents, err := os.ReadFile(fullFilePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	json.Unmarshal(fileContents, &config)
}
