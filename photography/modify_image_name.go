package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	prefix          = "IMG"
	myPrefixPattern = "I%v"
	targetSuffix    = []string{".jpg", ".JPG", ".CR2", ".MOV"}
)

func GetFileModTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	return fileInfo.ModTime()
}

func WalkDir(dirPath string) (files []string, err error) {
	fromSlash := filepath.FromSlash(dirPath)
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}
	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			subDirPath := filepath.Join(fromSlash, fi.Name())
			subFiles, _ := WalkDir(subDirPath)
			files = append(files, subFiles...)
		} else {
			ok := false
			if strings.HasPrefix(fi.Name(), prefix) {
				for _, suffix := range targetSuffix {
					if strings.HasSuffix(fi.Name(), suffix) {
						ok = true
						break
					}
				}
				if ok {
					files = append(files, dirPath+PthSep+fi.Name())
				}
			}
		}
	}
	return files, nil
}

func RenameFile(myFilePath, prefix, myPrefix string) {
	baseName := filepath.Base(myFilePath)
	if strings.HasPrefix(baseName, prefix) {
		baseName := filepath.Base(myFilePath)
		absName, _ := filepath.Abs(myFilePath)
		absDir := filepath.Dir(absName)
		newBaseName := strings.Replace(baseName, prefix, myPrefix, 1)
		newAbsName := filepath.Join(absDir, newBaseName)
		err := os.Rename(absName, newAbsName)
		if err != nil {
			fmt.Println(baseName + "->" + newBaseName + " Failed")
		} else {
			fmt.Println(baseName + "->" + newBaseName)
		}
	}
}

func RenameFileInBatch(dirPath string) {
	curUser, err := user.Current()
	if err != nil {
		return
	} else {
		if strings.HasPrefix(dirPath, "~") {
			dirPath = strings.Replace(dirPath, "~", curUser.HomeDir, 1)
		}
		if strings.HasPrefix(dirPath, ".") {
			pwd, _ := os.Getwd()
			dirPath = strings.Replace(dirPath, ".", pwd, 1)
		}
	}
	files, _ := WalkDir(dirPath)

	var wg sync.WaitGroup
	var sema = make(chan struct{}, 50)

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			sema <- struct{}{}
			defer func() { <-sema }()
			defer wg.Done()
			createTime := GetFileModTime(f) // how to get create time from unix?
			createTimeStr := createTime.Format("200601")
			myPrefix := strings.Replace(myPrefixPattern, "%v", createTimeStr, 1)
			RenameFile(f, prefix, myPrefix)
			//RenameFile(f, myPrefix, prefix)
		}(file)
	}
	wg.Wait()
}

func main() {
	var dirPath string
	if len(os.Args) >= 2 {
		dirPath = os.Args[1]
		RenameFileInBatch(dirPath)
	} else {
		fmt.Print("Enter dir path: ")
		_, err := fmt.Scanln(&dirPath)
		if err == nil {
			RenameFileInBatch(dirPath)
		}
	}
}
