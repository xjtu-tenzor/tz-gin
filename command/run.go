package command

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/cosmtrek/air/runner"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/semver"
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
	if err != nil {
		return nil, err
	}

	goVersion, err := getGoVersion()
	if err != nil {
		return nil, err
	}
	fmt.Println(*goVersion)

	if semver.Compare(*goVersion, "1.20.0") == -1 {
		if root != "." {
			return nil, errors.New("-d flag not support below go version 1.20, please upgrade your golang version")
		}
		if ok := checkWindows(); ok {
			cfg.Build.IncludeFile = []string{".env", "README.md"}
			cfg.Build.Cmd = fmt.Sprintf("go build -o ./tmp/main.exe")
			cfg.Build.Bin = fmt.Sprintf(".\\tmp\\main.exe")
		} else {
			cfg.Build.IncludeFile = []string{".env", "README.md"}
			cfg.Build.Cmd = fmt.Sprintf("go build -o ./tmp/main")
			cfg.Build.Bin = fmt.Sprintf("./tmp/main")
		}
		return cfg, nil
	}

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

func getGoVersion() (*string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	goVersion := parseGoVersion(string(output))
	if goVersion == "" {
		return &goVersion, errors.New("go version parse error")
	}
	return &goVersion, nil
}

func parseGoVersion(output string) string {
	fields := strings.Fields(output)
	if len(fields) >= 3 {
		return strings.TrimPrefix(fields[2], "go")
	}
	return ""
}
