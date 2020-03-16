package common

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// Gets the Shared libraries extension included by dot, related to current O/S
func GetShareLibExt() string {
	if runtime.GOOS == "windows" {
		return ".dll"
	}
	return ".so"
}

func GetCurrentPath() string {
	wd, err := os.Getwd()
	if err != nil {
		exec, err := os.Executable()
		if err != nil {
			return HomeFolder()
		}
		return filepath.Dir(exec)
	}
	return wd
}

func HomeFolder() string {
	usr, err := user.Current()
	if err != nil {
		return os.TempDir()
	}
	return usr.HomeDir
}

