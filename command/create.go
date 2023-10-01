package command

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/urfave/cli/v2"
	"github.com/xjtu-tenzor/tz-gin/util"
)

var directoryString *string
var projectName string

func prepareDirectory() error {
	info, err := os.Stat(*directoryString)

	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(*directoryString, 0755)
			if err != nil {
				// util.ErrMsg("Create directory failed")
				// os.Exit(1)
				return cli.Exit("failed to create directory", 1)
			}
			info, err = os.Stat(*directoryString)
			if err != nil {
				// cli.HandleExitCoder(errors.New("123"))
				// util.ErrMsg(err.Error())
				// os.Exit(1)
				return cli.Exit(err.Error(), 1)
			}
		} else {
			// util.ErrMsg(err.Error())
			// os.Exit(1)
			return cli.Exit(err.Error(), 1)
		}
	}

	if !info.IsDir() {
		return cli.Exit("the path you select is folder", 1)
		// util.ErrMsg("The path you select is not a folder")
		// os.Exit(1)
	}

	file, err := os.Open(*directoryString)
	if err != nil {
		if !os.IsNotExist(err) {
			// util.ErrMsg(err.Error())
			// os.Exit(1)
			return cli.Exit(err.Error(), 1)
		}
	}
	defer file.Close()

	names, err := file.Readdirnames(-1)
	if err != nil {
		// util.ErrMsg(err.Error())
		// os.Exit(1)
		return cli.Exit(err.Error(), 1)
	}

	if len(names) != 0 {
		var per string
		for {
			util.WarnMsg("The folder is not empty, are you sure you want to create project in this directory? [y/N] ")

			if _, err := fmt.Scanf("%s\n", &per); err != nil {
				if err.Error() == "unexpected newline" {
					return cli.Exit("Interrupt by none-empty folder", 1)
				}
				return cli.Exit(err.Error(), 1)
			}
			if per == "" || per == "N" || per == "n" {
				return cli.Exit("Interrupt by none-empty folder", 1)
			}
			if per == "y" || per == "Y" {
				break
			}
		}
	}

	for _, n := range names {
		os.RemoveAll(path.Join(*directoryString, n))
	}

	return nil
}

func parseParams(c *cli.Context) error {
	directoryString = new(string)
	*directoryString = c.String("directory")

	if len(*directoryString) == 0 {
		*directoryString = "./"
	}

	projectName = c.Args().First()

	if len(projectName) == 0 {
		fmt.Print("Input your project name: ")
		if _, err := fmt.Scanf("%s", &projectName); err != nil {
			return cli.Exit(err.Error(), 1)
		}
	}

	if *directoryString == "./" {
		*directoryString = path.Join(*directoryString, projectName)
	}

	return nil
}

func getDownlowdResponse(addr string) (*http.Response, error) {
	response, err := http.Get(addr)
	if err != nil {
		return nil, cli.Exit(err.Error(), 1)
	}

	if response.StatusCode != http.StatusOK {
		return nil, cli.Exit(fmt.Sprintf("http request failed with status code %d", response.StatusCode), 1)
	}

	return response, nil
}

func render(zipReader *zip.Reader) error {
	for _, file := range zipReader.File {
		destFilePath := filepath.Join(*directoryString, strings.TrimPrefix(file.Name, "tz-gin-template-master"))

		err := os.MkdirAll(filepath.Dir(destFilePath), os.ModePerm)
		if err != nil {
			return err
		}

		if !file.FileInfo().IsDir() {
			destFile, err := os.Create(destFilePath)
			if err != nil {
				return err
			}

			defer destFile.Close()

			zipFile, err := file.Open()
			if err != nil {
				return err
			}

			var buffer bytes.Buffer
			_, err = io.Copy(&buffer, zipFile)
			if err != nil {
				return err
			}

			stringReader := strings.NewReader(strings.Replace(buffer.String(), "template", projectName, -1))
			_, err = io.Copy(destFile, stringReader)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func Create(c *cli.Context) error {

	if err := parseParams(c); err != nil {
		return err
	}

	if err := prepareDirectory(); err != nil {
		return err
	}

	stop := make(chan int, 1)
	go util.Loading(stop)

	downloadResponse, err := getDownlowdResponse(c.String("remote"))

	if err != nil {
		return cli.Exit(err.Error(), 1)
	}
	defer downloadResponse.Body.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, downloadResponse.Body)

	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	err = render(zipReader)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	stop <- 1
	util.SuccessMsg(fmt.Sprintf("\nProject has been generate into folder: %s\nUse command as follow\n\tcd %s\n\tgo run .\n", *directoryString, *directoryString))

	return nil
}
