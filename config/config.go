package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/0x6666/backup/log"
	"github.com/BurntSushi/toml"
)

type Task struct {
	Name      string
	Src       string
	Skip      []string `toml:"files"`
	IsContent bool     `toml:"content"`
	Dest      string
	Sync      bool
	Log       bool `toml:"log2file"`
	Count     int
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
