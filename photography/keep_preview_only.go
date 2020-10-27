package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func fileNameWithoutExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func unique(slice []string) []string {
	encountered := map[string]int{}
	var diff []string

	for _, v := range slice {
		encountered[v] = encountered[v] + 1
	}

	for _, v := range slice {
		if encountered[v] == 1 {
			diff = append(diff, v)
		}
	}
	return diff
}

func deleteFilePath(filePaths []string) {
	for _, filePath := range filePaths {
		e := os.Remove(filePath)
		if e != nil {
			fmt.Println(e)
		}
	}
}

func analyzeFilesToDelete(dirPath string) ([]string, error) {
	var filesToKeep []string
	var filesToDelete []string
	var allFileName []string
	filePathMap := make(map[string]bool)
	dirPath = filepath.Clean(dirPath)
	fileInfos, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return filesToDelete, err
	}
	for _, fi := range fileInfos {
		filePathMap[fi.Name()] = true
	}
	for _, fi := range fileInfos {
		fiName := fi.Name()
		allFileName = append(allFileName, fiName)
		fiNameWithoutExt := fileNameWithoutExtension(fiName)
		switch {
		case strings.HasSuffix(fiName, "CR2"):
			if _, ok := filePathMap[fiNameWithoutExt+".JPG"]; !ok {
				filesToKeep = append(filesToKeep, fiName)
			}
		default:
			filesToKeep = append(filesToKeep, fiName)
		}
	}
	filesToDelete = unique(append(allFileName, filesToKeep...))
	var filePathToDelete []string
	for _, fileName := range filesToDelete {
		filePathToDelete = append(filePathToDelete, filepath.Join(dirPath, fileName))
	}
	return filePathToDelete, nil
}

func main() {
	var dirPath string
	if len(os.Args) >= 2 {
		dirPath = os.Args[1]
	} else {
		fmt.Print("Enter dir path: ")
		_, _ = fmt.Scanln(&dirPath)
	}
	filePathToDelete, err := analyzeFilesToDelete(dirPath)
	if err != nil {
		return
	}
	deleteFilePath(filePathToDelete)
}
