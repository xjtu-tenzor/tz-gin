package config

import (
	_ "embed"
	"strings"

	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
)

type Config struct {
	APP        *App   `mapstructure:"app"`
	RemoteAddr string `mapstructure:"remote_addr"`
}

type App struct {
	Name    string        `mapstructure:"name"`
	Usage   string        `mapstructure:"usage"`
	Authors []*cli.Author `mapstructure:"authors"`
	Version string        `mapstructure:"version"`
}

func New() *Config {
	return &Config{
		APP: &App{
			Name:  "tz-gin-cli",
			Usage: "quickly build tenzor normalizing go-gin code",
			Authors: []*cli.Author{
				{
					Name:  "tenzor/tiaozhan",
					Email: "contact@tiaozhan.com",
				},
			},
			Version: "0.2.0",
		},
	}
}

func (cfg *Config) Parse(config string) *Config {
	v := viper.New()
	v.SetConfigType("toml")
	v.ReadConfig(strings.NewReader(config))
	tmp := &Config{}
	if err := v.Unmarshal(tmp); err != nil {
		panic("failed to init api config, error: " + err.Error())
	}
	return tmp
}

func (cfg *Config) Load(app *cli.App, config string) {
	tmp := cfg.Parse(config)
	if tmp.APP != nil {
		if tmp.APP.Name != "" {
			cfg.APP.Name = tmp.APP.Name
		}
		if tmp.APP.Authors != nil {
			cfg.APP.Authors = tmp.APP.Authors
		}
		if tmp.APP.Usage != "" {
			cfg.APP.Usage = tmp.APP.Usage
		}
		if tmp.APP.Version != "" {
			cfg.APP.Version = tmp.APP.Version
		}
	}
	app.Name = cfg.APP.Name
	app.Version = cfg.APP.Version
	app.Usage = cfg.APP.Usage

	app.Authors = cfg.APP.Authors
}
