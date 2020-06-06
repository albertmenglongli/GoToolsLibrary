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

func GetFileModTime(path string) time.Time {
	fileInfo, _ := os.Stat(path)
	return fileInfo.ModTime()
}

func WalkDir(dirPath string) (files []string, dirs []string, err error) {
	targetSuffix := []string{".jpg", ".JPG", "CR2"}
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}
	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			// todo
		} else {
			ok := false
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
	return files, dirs, nil
}

func RenameFile(filePath, prefix, myPrefix string) {
	baseName := filepath.Base(filePath)
	PthSep := string(os.PathSeparator)
	if strings.HasPrefix(baseName, prefix) {
		absName, _ := filepath.Abs(filePath)
		dir := filepath.Dir(absName)
		newBaseName := strings.Replace(baseName, prefix, myPrefix, 1)
		newAbsName := dir + PthSep + newBaseName
		err := os.Rename(absName, newAbsName)
		if err != nil {
			fmt.Println(baseName + "->" + newBaseName + " Failed")
		} else {
			fmt.Println(baseName + "->" + newBaseName)
		}
	}
}

func RenameBatch(dirPath string) {
	curUser, err := user.Current()
	if err == nil {
		dirPath = strings.Replace(dirPath, "~", curUser.HomeDir, 1)
	} else {
		return
	}

	files, _, _ := WalkDir(dirPath)
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			createTime := GetFileModTime(f)
			createTimeStr := createTime.Format("200601")
			prefix := "IMG"
			myPrefix := "I" + createTimeStr
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
		RenameBatch(dirPath)
	} else {
		fmt.Print("Enter dir path: ")
		_, err := fmt.Scanln(&dirPath)
		if err == nil {
			RenameBatch(dirPath)
		}
	}

}
