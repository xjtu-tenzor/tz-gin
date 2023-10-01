package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/urfave/cli/v2"
	"github.com/xjtu-tenzor/tz-gin/util"
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
		cmd := exec.Command(path, "install ", fmt.Sprintf("github.com/xjtu-tenzor/tz-gin@%s", *ver))
		err := cmd.Start()
		if err != nil {
			return cli.Exit(err.Error(), 1)
		}
		util.SuccessMsg(fmt.Sprintf("\nSuccessfully update to %s\n", *ver))
		return nil
	}

	stop <- 1
	return err
}

func getLatestVer() (*string, error) {
	apiUrl := "https://api.github.com/repos/xjtu-tenzor/tz-gin/releases/latest"
	response, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var releaseInfo struct {
		TagName string `json:"tag_name"`
	}

	err = json.NewDecoder(response.Body).Decode(&releaseInfo)

	if err != nil {
		return nil, err
	}

	return &releaseInfo.TagName, nil
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
