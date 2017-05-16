package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"strings"

	"github.com/0x6666/backup/config"
	"github.com/0x6666/backup/log"
	"github.com/0x6666/backup/zip"
)

func cdCWD() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	log.SetLevel(log.LevelInfo | log.LevelDebug | log.LevelWarn | log.LevelError)
	defer log.Close()

	if err := cdCWD(); err != nil {
		log.Error(err.Error())
	}

	err := config.Initialize("./backup.toml")
	if err != nil {
		log.Error(err.Error())
		return
	}

	for idx, t := range config.GConfig.Task {
		if len(t.Name) == 0 {
			log.Error("task name [%v] is empty", idx)
			continue
		}

		if err = doTask(&t); err != nil {
			log.Error(err.Error())
		}
	}
}

const (
	nameFormat = "2006-01-02--15-04-05"
)

func newZipFilename(taskName string) string {
	return taskName + "-" + time.Now().Format(nameFormat) + ".zip"
}

func doTask(task *config.Task) error {

	log.Info("start task [%v]", task.Name)
	if len(task.Src) == 0 {
		log.Error("src path is empty")
		return errors.New("src path is empty")
	}

	if len(task.Dest) == 0 {
		log.Error("dest path is empty")
		return errors.New("dest path is empty")
	}

	//TODO: windows
	destFilePath := task.Dest
	fileInfo, err := os.Stat(destFilePath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(destFilePath, 0777)
		if err != nil {
			log.Error("create dir [%v] failed: %v", destFilePath, err.Error())
			return err
		}
		ch := destFilePath[len(destFilePath)-1]
		if ch != '/' && ch != '\\' {
			destFilePath = destFilePath + "/"
		}
		destFilePath = destFilePath + newZipFilename(task.Name)

	} else {
		if fileInfo.IsDir() {
			ch := destFilePath[len(destFilePath)-1]
			if ch != '/' && ch != '\\' {
				destFilePath = destFilePath + "/"
			}
			destFilePath = destFilePath + newZipFilename(task.Name)
		} else {
			//TODO: backup
			os.Remove(destFilePath)
		}
	}

	if task.Log {
		log.SetLogFile(destFilePath + ".log")
		defer log.SetLogFile("")
	}

	err = zip.ZipFolder(task.Src, destFilePath, task.Skip, task.IsContent)

	if err != nil {
		return err
	}

	return clean(task)
}

func clean(task *config.Task) error {

	if task.Sync == true || task.Count <= 0 {
		return nil
	}

	destPath := task.Dest
	fileInfo, err := os.Stat(destPath)
	if err != nil {
		log.Error("get dest file path info failed: %v", err)
		return err
	}

	if !fileInfo.IsDir() {
		return nil
	}

	files, err := listDir(destPath, task)
	if err != nil {
		return err
	}

	var lastTime time.Time

	for _, t := range files {
		if lastTime.Before(t) {
			lastTime = t
		}
	}

	for path, t := range files {
		if lastTime.Sub(t) < time.Hour*24*time.Duration(task.Count) {
			continue
		}

		if err := os.Remove(path); err != nil {
			log.Error("clean file [%v] failed: %v", path, err)
		}
	}

	return nil
}

func listDir(dirPath string, task *config.Task) (map[string]time.Time, error) {
	files := map[string]time.Time{}

	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Error("ReadDir [%v] failed: %v", dirPath, err)
		return nil, err
	}

	taskName := task.Name

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}

		name := fi.Name()
		if strings.HasSuffix(name, ".zip.log") {
			t, err := time.Parse(taskName+"-"+nameFormat+".zip.log", name)
			if err != nil {
				log.Error("parse file name [%v] failed: %v", name, err)
				continue
			}
			files[dirPath+PthSep+name] = t
			continue
		}

		if strings.HasSuffix(name, ".zip") {
			t, err := time.Parse(taskName+"-"+nameFormat+".zip", name)
			if err != nil {
				log.Error("parse file name [%v] failed: %v", name, err)
				continue
			}
			files[dirPath+PthSep+name] = t
		}
	}
	return files, nil
}
