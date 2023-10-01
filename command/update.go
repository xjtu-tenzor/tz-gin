package command

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/urfave/cli/v2"
)

func Update(c *cli.Context) error {
	path, err := checkExists()
	if len(path) != 0 && err != nil {
		cmd := exec.Command(path, "install ", "github.com/xjtu-tenzor/tz-gin", "@latest")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalln("fail ", err)
			return err
		}
		fmt.Printf("update out :\n%s\n", string(out))
		return nil
	}
	return err
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
