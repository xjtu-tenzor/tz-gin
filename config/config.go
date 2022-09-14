package config

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

type Config struct {
	APP *App `mapstructure:"app"`
}

type App struct {
	name    string `mapstructure:"name"`
	usage   string `mapstructure:"usage"`
	author  string `mapstructure:"author"`
	version string `mapstructure:"version"`
}

func New() *Config {
	return &Config{
		APP: &App{
			name:    "tz-gin-cli",
			usage:   "quickly build tenzor normalizing go-gin code",
			author:  "tenzor/tiaozhan",
			version: "0.2.0",
		},
	}
}

func (cfg *Config) Load(key string, app *cli.App) {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config file, error: " + err.Error())
	}
	tmp := &Config{}
	if err := viper.UnmarshalKey(key, tmp); err != nil {
		panic("failed to init api config, error: " + err.Error())
	}
	if tmp.APP != nil {
		if tmp.APP.name != "" {
			cfg.APP.name = tmp.APP.name
		}
		if tmp.APP.author != "" {
			cfg.APP.author = tmp.APP.author
		}
		if tmp.APP.usage != "" {
			cfg.APP.usage = tmp.APP.usage
		}
		if tmp.APP.version != "" {
			cfg.APP.version = tmp.APP.version
		}
	}
	app.Name = cfg.APP.name
	app.Version = cfg.APP.version
	app.Usage = cfg.APP.usage
	app.Author = cfg.APP.author
}
