package io

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// Check existance of a given file by name
func ExistsFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Retrieve current wotrking folder
func GetCurrentFolder() string {
	cPath, err := os.Getwd()
	if err != nil {
		cPath, err := os.Executable()
		if err != nil {
			cPath, err = filepath.Abs(".")
			if err != nil {
				return "."
			}
			return cPath
		}
		cPath = path.Dir(cPath)
	}
	cPath, err = filepath.Abs(cPath)
	if err != nil {
		return "."
	}
	return filepath.Clean(cPath)
}

// Gets the Path Separator as string type
func GetPathSeparator() string {
	if runtime.GOOS == "windows" {
		return "\\"
	}
	return string(os.PathSeparator)
}

//Verifies if a atring file path corresponds to a directory
func IsFolder(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// Gets files in a folder (eventually recursively)
func GetFiles(path string, recursive bool) []string {
	var out []string = make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return out
	}
	for _, file := range files {
		var name = path + GetPathSeparator() + file.Name()
		if !file.IsDir() {
			out = append(out, name)
		} else if recursive {
			var filesX []string = GetFiles(name, recursive)
			for _, fileX := range filesX {
				out = append(out, fileX)
			}
		}
	}
	return out
}

// Gets files in a folder (eventually recursively), which name matches with given function execution
func GetMatchedFiles(path string, recursive bool, matcher func(string) bool) []string {
	var out []string = make([]string, 0)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return out
	}
	for _, file := range files {
		var name = path + GetPathSeparator() + file.Name()
		if !file.IsDir() {
			if matcher(name) {
				out = append(out, name)
			}
		} else if recursive {
			var filesX []string = GetMatchedFiles(name, recursive, matcher)
			for _, fileX := range filesX {
				if matcher(name) {
					out = append(out, fileX)
				}
			}
		}
	}
	return out
}

// Finds files recursively or not, in a given path folder, with a file name prefix token
func FindFilesIn(path string, recursive bool, searchText string) []string {
	return GetMatchedFiles(path, recursive, func(name string) bool {
		return strings.Contains(name, searchText)
	})
}
