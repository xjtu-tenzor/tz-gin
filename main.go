package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

var directoryString *string
var projectName string

//go:embed template.tar.gz
var template []byte

func init() {
	flag.Usage = func() {
		fmt.Print("Use")
		fmt.Printf(" \033[0;32;40m%s\033[0m", "create-gin <project-dir>")
		fmt.Print(" to create tz-gin project in the directory")
		fmt.Printf(" \033[0;32;40m%s\033[0m\n", "<project-dir>")
	}

	directoryString = flag.String("d", "./", "Input the directory you want to create to.")

	flag.Parse()
}

func parseParams() {
	if flag.NArg() > 1 {
		errMsg("Only one project name can be provided!")
		os.Exit(1)
	}

	if flag.NArg() != 0 {
		projectName = flag.Args()[0]
	}

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
			warnMsg("The folder is not emtpy, are you sure you want to create project in this directory? [y/N]")
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

func main() {
	successMsg("Welcome to use this cli. Developed by tenzor.")

	parseParams()

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

}
