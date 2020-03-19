package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Path string
}

func LoadConfig(cfg string) {
	logrus.SetOutput(os.Stdout)
	logFmt := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	logrus.SetFormatter(logFmt)

	c := &Config{Path: cfg}

	// 初始化配置文件
	if err := c.initConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			"module": "config",
			"method": "init",
		}).Panicf("配置文件读取失败", err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"module": "config",
		"method": "init",
	}).Info("配置文件读取成功")

	// 监听配置文件是否改变
	c.watchConfig()
}

func (c *Config) initConfig() error {
	if c.Path != "" {
		// 如果指定了配置文件, 则解析指定的配置文件
		viper.SetConfigFile(c.Path)
	} else {
		// 如果没有指定配置文件,则解析默认的配置文件
		path, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		configPath := filepath.Join(path, "/config")
		viper.AddConfigPath(configPath)
		viper.SetConfigName("local")
	}

	viper.SetConfigType("yaml")
	// viper 解析配置文件
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (c *Config) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(event fsnotify.Event) {
		log.Printf("config file changed: %s\n", event.Name)
	})
}
