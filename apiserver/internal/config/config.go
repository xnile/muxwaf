package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/xnile/muxwaf/pkg/logx"
)

type Config struct {
	Name string
}

func (c *Config) initConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("./conf")
		viper.SetConfigFile("config")
	}

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MUXWAF")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	//fmt.Println("vvvvv:", viper.GetString("postgresql.host"))

	return nil
}

func (c *Config) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		//log.Printf("Config file changed: %s", e.Name)
	})
}

func (c *Config) initLog() {
	config := logx.Config{
		Writers:        viper.GetString("log.writers"),
		LoggerLevel:    viper.GetString("log.logger_level"),
		LoggerFile:     viper.GetString("log.logger_file"),
		LogBackupCount: viper.GetInt("log.log_backup_count"),
	}
	err := logx.NewLogger(&config, logx.InstanceZapLogger)
	if err != nil {
		fmt.Printf("InitWithConfig err: %v", err)
	}
}

func Init(cfg string) error {
	c := Config{cfg}

	if err := c.initConfig(); err != nil {
		return err
	}

	c.initLog()
	c.watchConfig()

	return nil
}
