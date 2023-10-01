package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"sort"

	"github.com/urfave/cli/v2"
	"github.com/xjtu-tenzor/tz-gin/util"
	"golang.org/x/mod/semver"
)

func Update(c *cli.Context) error {
	stop := make(chan int, 1)
	go util.Loading(stop)
	path, err := checkExists()
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	ver, err := getLatestVer()
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	if len(path) != 0 && err == nil {
		cmd := exec.Command(path, "install", fmt.Sprintf("github.com/xjtu-tenzor/tz-gin@%s", *ver))

		_, err := cmd.StdoutPipe()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
		_, err = cmd.StderrPipe()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		err = cmd.Start()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		err = cmd.Wait()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}

		stop <- 1
		util.SuccessMsg(fmt.Sprintf("\nSuccessfully update to %s\n", *ver))
		return nil
	}

	return err
}

func getLatestVer() (*string, error) {
	apiUrl := "https://api.github.com/repos/xjtu-tenzor/tz-gin/tags"
	response, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch tags")
	}

	var releaseInfo []struct {
		Name string `json:"name"`
	}

	err = json.NewDecoder(response.Body).Decode(&releaseInfo)

	if err != nil {
		return nil, err
	}

	sort.Slice(releaseInfo, func(i, j int) bool {
		return semver.Compare(releaseInfo[i].Name, releaseInfo[j].Name) == 1
	})

	return &releaseInfo[0].Name, nil
}

func checkExists() (string, error) {
	path, err := exec.LookPath("go")
	if err != nil {
		fmt.Printf("cannot find command\"go\"")
		return "", err
	} else {
		// fmt.Printf("\"go\" executable is in '%s'\n", path)
		return path, nil
	}
}
