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
	name    string        `mapstructure:"name"`
	usage   string        `mapstructure:"usage"`
	authors []*cli.Author `mapstructure:"authors"`
	version string        `mapstructure:"version"`
}

func New() *Config {
	return &Config{
		APP: &App{
			name:  "tz-gin-cli",
			usage: "quickly build tenzor normalizing go-gin code",
			authors: []*cli.Author{
				{
					Name:  "tenzor/tiaozhan",
					Email: "contact@tiaozhan.com",
				},
			},
			version: "0.2.0",
		},
	}
}

func (cfg *Config) Parse(key string, config string) *Config {
	v := viper.New()
	v.SetConfigType("toml")
	v.ReadConfig(strings.NewReader(config))
	tmp := &Config{}
	if err := v.UnmarshalKey(key, tmp); err != nil {
		panic("failed to init api config, error: " + err.Error())
	}
	return tmp
}

func (cfg *Config) Load(key string, app *cli.App, config string) {
	tmp := cfg.Parse(key, config)
	if tmp.APP != nil {
		if tmp.APP.name != "" {
			cfg.APP.name = tmp.APP.name
		}
		if tmp.APP.authors != nil {
			cfg.APP.authors = tmp.APP.authors
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

	app.Authors = cfg.APP.authors
}
