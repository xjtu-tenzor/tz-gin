package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/urfave/cli"
	"github.com/xjtu-tenzor/tz-gin/config"
)

var directoryString *string
var projectName string

//go:embed template.tar.gz
var template []byte

func parseParams(c *cli.Context) {
	directoryString = flag.String("d", "./", "Input the directory you want to create to.")

	projectName = c.Args().First()

	if len(projectName) == 0 {
		fmt.Print("Input your project name : ")
		if _, err := fmt.Scanf("%s", &projectName); err != nil {
			errMsg(err.Error())
			os.Exit(1)
		}
	}

	if *directoryString == "./" {
		*directoryString = path.Join(*directoryString, projectName)
	}
}

func prepareDirectory() {
	info, err := os.Stat(*directoryString)

	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(*directoryString, 0755)
			if err != nil {
				errMsg("Create directory failed")
				os.Exit(1)
			}
			info, err = os.Stat(*directoryString)
			if err != nil {
				errMsg(err.Error())
				os.Exit(1)
			}
		} else {
			errMsg(err.Error())
			os.Exit(1)
		}
	}

	if !info.IsDir() {
		errMsg("The path you select is not a folder")
		os.Exit(1)
	}

	file, err := os.Open(*directoryString)
	if err != nil {
		if !os.IsNotExist(err) {
			errMsg(err.Error())
			os.Exit(1)
		}
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		errMsg(err.Error())
		os.Exit(1)
	}

	if len(names) != 0 {
		var per string
		for {
			warnMsg("The folder is not empty, are you sure you want to create project in this directory? [y/N]")
			if _, err := fmt.Scanf("%s\n", &per); err != nil {
				if err.Error() == "unexpected newline" {
					warnMsg("Interrupt by none-empty folder")
					os.Exit(1)
				}
				errMsg(err.Error())
				os.Exit(1)
			}
			if per == "" || per == "N" || per == "n" {
				warnMsg("Interrupt by none-empty folder")
				os.Exit(1)
			}
			if per == "y" || per == "Y" {
				break
			}
		}
	}

	for _, n := range names {
		os.RemoveAll(path.Join(*directoryString, n))
	}
}

func create(c *cli.Context) error {
	successMsg("Welcome to use this cli. Developed by tenzor.")

	parseParams(c)

	prepareDirectory()

	gzipReader, err := gzip.NewReader(bytes.NewReader(template))
	if err != nil {
		errMsg("Internal error: open template error " + err.Error())
		os.Exit(1)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	for {
		h, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			errMsg("Internal error: tar error")
			os.Exit(1)
		}

		destName := path.Join(*directoryString, h.Name)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(destName, 0755); err != nil {
				errMsg("Internal error: extract error " + err.Error())
				os.Exit(1)
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(destName, os.O_CREATE|os.O_WRONLY, 0666)
			var tempBuffer bytes.Buffer
			io.Copy(&tempBuffer, tarReader)

			stringReader := strings.NewReader(strings.Replace(tempBuffer.String(), "template", projectName, -1))
			if err != nil {
				errMsg("Internal error: extract error " + err.Error())
				os.Exit(1)
			}
			_, err = io.Copy(outFile, stringReader)
			if err != nil {
				errMsg("Internal error: writing error")
				os.Exit(1)
			}
			defer outFile.Close()
		}
		if err != nil {
			errMsg("Internal error: extract error " + err.Error())
			os.Exit(1)
		}
	}
	return nil
}

func update(c *cli.Context) error {
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

func main() {

	cfg := config.New()

	app := cli.NewApp()
	cfg.Load("tz.gin", app)
	app.Commands = []cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "create operations ",
			Action: func(c *cli.Context) error {
				err := create(c)
				if err != nil {
					fmt.Println(err)
					return err
				}
				return nil
			},
		},
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "update operations ",
			Action: func(c *cli.Context) error {
				err := update(c)
				if err != nil {
					fmt.Println(err)
					return err
				}
				return nil
			},
		},
	}
	_ = app.Run(os.Args)
}
