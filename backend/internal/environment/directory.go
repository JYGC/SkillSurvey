package environment

import (
	"os"
	"path/filepath"
)

func AttachToExecutableDir(fileName string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exeDir := filepath.Dir(exe)
	fullFilePath := filepath.Join(exeDir, fileName)
	return fullFilePath
}
