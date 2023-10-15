package command

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/cosmtrek/air/runner"
	"github.com/urfave/cli/v2"
)

func Run(ctx *cli.Context) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var err error
	cfg, err := initConfig(ctx.String("directory"))
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	// cmdArgs := runner.ParseConfigFlag()
	// cfg.WithArgs(cmdArgs)
	r, err := runner.NewEngineWithConfig(cfg, debugMode)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	go func() {
		<-sigs
		r.Stop()
	}()

	defer func() {
		if e := recover(); e != nil {
			log.Fatalf("PANIC: %+v", e)
		}
	}()

	r.Run()

	return nil
}

func initConfig(root string) (*runner.Config, error) {
	cfg, err := runner.InitConfig("")
	cfg.Root = root
	if ok := checkWindows(); ok {
		cfg.Build.IncludeFile = []string{".env", "README.md"}
		cfg.Build.Cmd = fmt.Sprintf("go build -o ./tmp/main.exe -C %s", root)
		cfg.Build.Bin = fmt.Sprintf("%s\\tmp\\main.exe", root)
	} else {
		cfg.Build.IncludeFile = []string{".env", "README.md"}
		cfg.Build.Cmd = fmt.Sprintf("go build -o ./tmp/main -C %s", root)
		cfg.Build.Bin = fmt.Sprintf("%s/tmp/main", root)
	}

	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func checkWindows() bool {
	switch os := runtime.GOOS; os {
	case "windows":
		return true
	default:
		return false
	}
}
