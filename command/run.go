package command

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/urfave/cli/v2"
)

func Run(ctx *cli.Context) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	return nil
}

func checkWindows() bool {
	return runtime.GOOS == "windows"
}
