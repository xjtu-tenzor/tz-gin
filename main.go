package main

import (
	_ "embed"
	"os"

	"github.com/xjtu-tenzor/tz-gin/app"
	"github.com/xjtu-tenzor/tz-gin/util"
)

//go:embed config.toml
var configStirng string

func main() {
	app := app.InitApp(configStirng)
	util.SuccessMsg("Welcome to use this cli. Developed by tenzor.\n")

	app.Run(os.Args)

}
