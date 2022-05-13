package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/odpf/compass/cmd"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	command := cmd.New()

	if err := command.Execute(); err != nil {
		if strings.HasPrefix(err.Error(), "unknown command") {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(command.UsageString())
			os.Exit(exitOK)
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitError)
		}
	}
}
