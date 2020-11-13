package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	filesChan       = make(chan string, 50)
)

func GetFileModTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	return fileInfo.ModTime()
}

func WalkDirEntry(filePath string) {
	var wg sync.WaitGroup
	wg.Add(1)
	go WalkDir(filePath, filesChan, &wg)
	wg.Wait()
	close(filesChan)

}

func WalkDir(filePath string, files chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	filePath = filepath.Clean(filePath)
	dirPath := filepath.FromSlash(filePath)
	fileInfos, err := ioutil.ReadDir(filePath)
	if err != nil {
		return
	}

	for _, fi := range fileInfos {
		fiPath := filepath.Join(dirPath, fi.Name())
		if fi.IsDir() {
			wg.Add(1)
			go WalkDir(fiPath, files, wg)
		} else {
			if strings.HasPrefix(fi.Name(), prefix) {
				for _, suffix := range targetSuffix {
					if strings.HasSuffix(fi.Name(), suffix) {
						files <- fiPath
						break
					}
				}
			}
		}
	}
}

func RenameFile(myFilePath, prefix, myPrefix string) {
	baseName := filepath.Base(myFilePath)
	if strings.HasPrefix(baseName, prefix) {
		baseName := filepath.Base(myFilePath)
		absFilePath, _ := filepath.Abs(myFilePath)
		absDir := filepath.Dir(absFilePath)
		newBaseName := strings.Replace(baseName, prefix, myPrefix, 1)
		newAbsFilePath := filepath.Join(absDir, newBaseName)
		err := os.Rename(absFilePath, newAbsFilePath)
		if err != nil {
			RenameAfterChflags(myFilePath, prefix, myPrefix)
		} else {
			fmt.Println(baseName + "->" + newBaseName)
		}
	}
}

func RenameAfterChflags(myFilePath, prefix, myPrefix string) {
	baseName := filepath.Base(myFilePath)
	absFilePath, _ := filepath.Abs(myFilePath)
	absDir := filepath.Dir(absFilePath)
	newBaseName := strings.Replace(baseName, prefix, myPrefix, 1)
	newAbsFilePath := filepath.Join(absDir, newBaseName)
	err := os.Rename(absFilePath, newAbsFilePath)
	cmd := exec.Command("chflags", "nouchg", absFilePath)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
	} else {
		err = os.Rename(absFilePath, newAbsFilePath)
		if err != nil {
			fmt.Println(baseName + "->" + newBaseName + " Failed")
		} else {
			cmd := exec.Command("chflags", "uchg", newAbsFilePath)
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(baseName + "->" + newBaseName)
			}
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
	go WalkDirEntry(dirPath)

	var wg sync.WaitGroup
	var sema = make(chan struct{}, 100)
	for file := range filesChan {
		wg.Add(1)
		go func(f string) {
			sema <- struct{}{}
			defer func() { <-sema }()
			defer wg.Done()
			createTime := GetFileModTime(f) // how to get create time from unix?
			createTimeStr := createTime.Format("200601")
			myPrefix := strings.Replace(myPrefixPattern, "%v", createTimeStr, 1)
			RenameFile(f, prefix, myPrefix)
		}(file)
	}
	wg.Wait()
}

func main() {
	var dirPath string
	if len(os.Args) >= 2 {
		dirPath = os.Args[1]
	} else {
		fmt.Print("Enter dir path: ")
		_, _ = fmt.Scanln(&dirPath)
	}
	RenameFileInBatch(dirPath)
}
