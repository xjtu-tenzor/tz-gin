package main

import (
	"fmt"
)

func errMsg(msg string) {
	fmt.Printf("\033[0;31;40m%s\033[0m\n", msg)
}

func warnMsg(msg string) {
	fmt.Printf("\033[0;33;40m%s\033[0m\n", msg)
}

func successMsg(msg string) {
	fmt.Printf("\033[0;32;40m%s\033[0m\n", msg)
}
