package util

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora/v3"
)

func ErrMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Fprintf(os.Stderr, "%s", aurora.Red(msg))
}

func WarnMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Printf("%s", aurora.Yellow(msg))
}

func SuccessMsg(msg string) {
	aurora := aurora.NewAurora(true)
	fmt.Printf("%s", aurora.Green(msg))
}
