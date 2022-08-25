package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Print("Use")
		fmt.Printf(" \033[0;32;40m%s\033[0m", "create-gin <project-dir>")
		fmt.Print(" to create tz-gin project in the directory")
		fmt.Printf(" \033[0;32;40m%s\033[0m\n", " <project-dir>")
	}

	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

}
