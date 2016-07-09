package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/inimei/backup/config"
	"github.com/inimei/backup/log"
	"github.com/inimei/backup/zip"
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

func newLogFilename(taskName string) string {
	return taskName + "-" + time.Now().Format(nameFormat) + ".log"
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

	//todo
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
			//todo
			os.Remove(destFilePath)
		}
	}

	if task.Log {
		log.SetLogFile(destFilePath + ".log")
		defer log.SetLogFile("")
	}

	return zip.ZipFolder(task.Src, destFilePath, task.Skip)
}
