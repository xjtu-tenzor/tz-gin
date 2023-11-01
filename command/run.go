package command

import (
	"context"
	"crypto/sha1"
	"crypto/subtle"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v2"
	"github.com/xjtu-tenzor/tz-gin/util"
	"golang.org/x/mod/semver"
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
var support bool
var directory string

func Run(ctx *cli.Context) error {
	util.SuccessMsg("[info] press ctrl-c or c to exit, r to force rebuild\n")
	windows = checkWindows()
	var err error
	support, err = cSupport()
	if err != nil {
		return cli.Exit(err, 1)
	}
	directory = ctx.String("directory")
	if !support && directory != "." {
		return cli.Exit(errors.New("go version below 1.20.0 do not support -d flag"), 1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	buildTrigger := make(chan struct{})
	chanErr := make(chan error)

	defer clean()

	cancelCtx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()

	wg.Add(3)
	go watcherRoutine(cancelCtx, &wg, buildTrigger, chanErr)
	go keyRoutine(cancelCtx, &wg, buildTrigger, sigs, chanErr)
	go buildRoutine(cancelCtx, &wg, buildTrigger, chanErr)

	buildTrigger <- struct{}{}

	select {
	case <-sigs:
		return nil
	case err := <-chanErr:
		return cli.Exit(err, 1)
	}
}

func watcherRoutine(ctx context.Context, wg *sync.WaitGroup, buildTrigger chan struct{}, chanErr chan error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		chanErr <- err
		return
	}

	defer watcher.Close()

	err = watcher.Add(directory)
	if err != nil {
		chanErr <- err
		return
	}

	for _, file := range fileToWatch {
		err = watcher.Add(path.Join(directory, file))
		if err != nil {
			chanErr <- err
			return
		}
	}
	fileSha := map[string][]byte{}
	defer wg.Done()
	for {
		select {
		case event := <-watcher.Events:
			if strings.HasPrefix(event.Name, "tmp") || strings.HasPrefix(event.Name, ".git") {
				continue
			}
			if event.Op.Has(fsnotify.Chmod) || event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Remove) || event.Op == 0 {
				continue
			}
			util.SuccessMsg("[watcher] event: " + event.String() + "\n")
			time.Sleep(100 * time.Millisecond)
			sha1hash, _ := getFileSha1(event.Name)
			if fileSha[event.Name] != nil && subtle.ConstantTimeCompare(fileSha[event.Name][:], sha1hash[:]) == 1 {
				util.WarnMsg("[watcher] file not change: " + event.Name + "\n")
				continue
			}
			fileSha[event.Name] = sha1hash
			buildTrigger <- struct{}{}
		case err := <-watcher.Errors:
			if err != nil {
				util.ErrMsg("[watcher] error: " + err.Error() + "\n")
			}
		case <-ctx.Done():
			return
		}
	}
}

func keyRoutine(ctx context.Context, wg *sync.WaitGroup, buildTrigger chan struct{}, sigs chan os.Signal, chanErr chan error) {
	err := keyboard.Open()
	defer keyboard.Close()
	if err != nil {
		chanErr <- err
		return
	}
	if err != nil {
		chanErr <- err
		return
	}
	for {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			run, key, err := keyboard.GetKey()
			if err != nil {
				util.WarnMsg("[key] error: " + err.Error() + "\n")
				continue
			}

			if key == keyboard.KeyCtrlC {
				sigs <- syscall.SIGINT
				wg.Done()
				return
			}

			if run == 'r' {
				buildTrigger <- struct{}{}
			}
			if run == 'c' {
				sigs <- syscall.SIGINT
				wg.Done()
				return
			}
		}
	}
}

func buildRoutine(ctx context.Context, wg *sync.WaitGroup, buildTrigger chan struct{}, chanErr chan error) {
	defer wg.Done()
	var execCmd *exec.Cmd
	for {
		select {
		case <-buildTrigger:
			util.SuccessMsg("[builder] building ...\n")

			var cmd *exec.Cmd
			if !windows {
				if support {
					cmd = exec.Command("go", "build", "-C", directory, "-o", "tmp/main", ".")
				} else {
					cmd = exec.Command("go", "build", "-o", "tmp/main", ".")
				}
			} else {
				if support {
					cmd = exec.Command("go", "build", "-C", directory, "-o", "tmp\\main.exe", ".")
				} else {
					cmd = exec.Command("go", "build", "-o", "tmp\\main.exe", ".")
				}
			}

			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			err := cmd.Run()
			if err != nil {
				util.ErrMsg("[builder]: build failed: " + err.Error() + "\n")
				continue
			}

			util.SuccessMsg("[builder] build finished\n")

			if execCmd != nil && execCmd.ProcessState != nil {
				util.WarnMsg("[runner] killing ...\n")
				err := execCmd.Process.Kill()
				if err != nil {
					chanErr <- err
					return
				}
				execCmd.Wait()
			}

			// exec
			if !windows {
				if support {
					cmd = exec.Command(path.Join(directory, "tmp/main"))
				} else {
					cmd = exec.Command("tmp/main")
				}
			} else {
				if support {
					cmd = exec.Command(path.Join(directory, "tmp\\main.exe"))
				} else {
					cmd = exec.Command("tmp\\main.exe")
				}
			}
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			if err := cmd.Start(); err != nil {
				chanErr <- err
				return
			}
			execCmd = cmd
			util.SuccessMsg("[runner] running ...\n")

		case <-ctx.Done():
			if execCmd != nil && execCmd.ProcessState != nil {
				util.WarnMsg("[runner] killing ...\n")
				execCmd.Process.Kill()
				execCmd.Wait()
			}
			return
		}

	}
}

func clean() error {
	err := os.RemoveAll(path.Join(directory, "tmp"))
	if err != nil {
		util.ErrMsg("[clean]: " + err.Error())
	}
	return err
}

func checkWindows() bool {
	return runtime.GOOS == "windows"
}

func getFileSha1(file string) ([]byte, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer fp.Close()

	sha1hash := sha1.New()
	_, err = io.Copy(sha1hash, fp)

	if err != nil {
		return nil, err
	}

	return sha1hash.Sum(nil), nil
}

func getGoVer() (string, error) {
	cmd := exec.Command("go", "version")
	bytes, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func cSupport() (bool, error) {
	goVer, err := getGoVer()
	if err != nil {
		return true, err
	}
	goVer = "v" + strings.Replace(strings.Split(goVer, " ")[2], "go", "", -1)

	return semver.Compare(goVer, "v1.20.0") != -1, nil
}
