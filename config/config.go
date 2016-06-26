package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/inimei/backup/log"
)

type Task struct {
	Name string
	Src  string
	Skip []string
	Dest string
	Sync bool
}

type Config struct {
	Task []Task
}

var GConfig Config

func Initialize(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "backup.toml"
	}

	f, err := os.Open(configFilePath)
	if err != nil {
		cwd, _ := os.Getwd()
		log.Error("os.Stat fail, %s ,please ensure account.conf exist. account.conf path:%s, cwd:%s", err, configFilePath, cwd)
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error("read config file error, %s", err)
		return err
	}

	if err := toml.Unmarshal(buf, &GConfig); err != nil {
		log.Error("unmarshal config failed, %s", err)
		return err
	}

	if len(GConfig.Task) == 0 {
		log.Error("No task!")
		return errors.New("No task!")
	}

	return nil
}
