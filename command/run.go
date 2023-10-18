package command

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
	"github.com/xjtu-tenzor/tz-gin/util"
)

var fileToWatch []string = []string{
	"config",
	"controller",
	"middleware",
	"model",
	"router",
	"service",
	"service/validator",
}

var windows bool

func Run(ctx *cli.Context) error {
	windows = checkWindows()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	buildTrigger := make(chan struct{})
	execTrigger := make(chan struct{})
	chanErr := make(chan error)

	cancelCtx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()

	directory := ctx.String("directory")

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return cli.Exit(err, 1)
	}

	defer watcher.Close()

	err = watcher.Add(directory)
	if err != nil {
		return cli.Exit(err, 1)
	}

	for _, file := range fileToWatch {
		err = watcher.Add(path.Join(directory, file))
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	wg.Add(3)
	go watcherRoutine(cancelCtx, &wg, watcher, buildTrigger, chanErr)
	go buildRoutine(cancelCtx, &wg, buildTrigger, execTrigger, chanErr)
	go execRoutine(cancelCtx, &wg, execTrigger, chanErr)

	buildTrigger <- struct{}{}

	select {
	case <-sigs:
		return nil
	case err := <-chanErr:
		return cli.Exit(err, 1)
	case <-buildTrigger:
		return nil
	}
}

func watcherRoutine(ctx context.Context, wg *sync.WaitGroup, watcher *fsnotify.Watcher, buildTrigger chan struct{}, chanErr chan error) {
	for {
		select {
		case event := <-watcher.Events:
			if strings.HasPrefix(event.Name, "tmp") {
				continue
			}
			// if event.Op == 0 {
			// 	wg.Done()
			// 	return
			// }
			util.SuccessMsg("[watcher] event: " + event.String() + "\n")
			if event.Has(fsnotify.Write) {
				buildTrigger <- struct{}{}
			}
		case err := <-watcher.Errors:
			if err != nil {
				util.ErrMsg("[watcher] error: " + err.Error() + "\n")
			}
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func buildRoutine(ctx context.Context, wg *sync.WaitGroup, buildTrigger chan struct{}, execTrigger chan struct{}, chanErr chan error) {
	for {
		select {
		case <-buildTrigger:
			util.SuccessMsg("[builder] building ...\n")

			var cmd *exec.Cmd
			if !windows {
				cmd = exec.Command("go", "build", "-o", "tmp/main", ".")
			} else {

				cmd = exec.Command("go", "build", "-o", "tmp\\main.exe", ".")
			}

			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			err := cmd.Run()
			if err != nil {
				util.ErrMsg(err.Error())
				chanErr <- err
			}

			util.SuccessMsg("[builder] build finished\n")
			execTrigger <- struct{}{}
		case <-ctx.Done():
			wg.Done()
			return
		}

	}
}

func execRoutine(ctx context.Context, wg *sync.WaitGroup, execTrigger chan struct{}, chanErr chan error) {
	var process *os.Process
	for {
		select {
		case <-ctx.Done():
			if process != nil {
				util.SuccessMsg("[runner] killing ...\n")
				process.Kill()
			}
			wg.Done()
			return
		case <-execTrigger:
			if process != nil {
				util.SuccessMsg("[runner] killing ...\n")
				if err := process.Kill(); err != nil {
					chanErr <- err
					return
				}
			}
			util.SuccessMsg("[runner] running ...\n")

			var cmd *exec.Cmd
			if !windows {
				cmd = exec.Command("tmp/main")
			} else {
				cmd = exec.Command("tmp\\main.exe")
			}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			if err := cmd.Start(); err != nil {
				chanErr <- err
				return
			}
			process = cmd.Process
		}
	}
}

func checkWindows() bool {
	return runtime.GOOS == "windows"
}
