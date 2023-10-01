package app

import (
	"github.com/xjtu-tenzor/tz-gin/command"
	"github.com/xjtu-tenzor/tz-gin/config"
	"github.com/xjtu-tenzor/tz-gin/util"

	"github.com/urfave/cli/v2"

	_ "embed"
)

func InitApp(configString string) *cli.App {
	cfg := config.New()

	app := cli.NewApp()
	cfg.Load("tz.gin", app, configString)
	cfgStruct := cfg.Parse("tz.gin", configString)

	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create operations",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "directory",
					Aliases: []string{"d"},
					Usage:   "the direcotory you want to generate",
				},
				&cli.StringFlag{
					Name:    "remote",
					Aliases: []string{"r"},
					Usage:   "the remote address you want to pull",
				},
			},
			Action: func(ctx *cli.Context) error {
				if err := ctx.Set("remote", cfgStruct.RemoteAddr); err != nil {
					return err
				}
				return command.Create(ctx)
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update operations",
			Action:  command.Update,
		},
	}

	app.ExitErrHandler = func(cCtx *cli.Context, err error) {
		if err != nil {
			util.ErrMsg(err.Error() + "\n")
		}
	}

	return app
}
