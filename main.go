package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/goto/compass/cli"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	cliConfig, err := cli.LoadConfig()
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if cmd, err := cli.New(cliConfig).ExecuteContextC(ctx); err != nil {
		printError(err)

		cmdErr := strings.HasPrefix(err.Error(), "unknown command")
		flagErr := strings.HasPrefix(err.Error(), "unknown flag")
		sflagErr := strings.HasPrefix(err.Error(), "unknown shorthand flag")

		if cmdErr || flagErr || sflagErr {
			if !strings.HasSuffix(err.Error(), "\n") {
				fmt.Println()
			}
			fmt.Println(cmd.UsageString())
			os.Exit(exitOK)
		} else {
			os.Exit(exitError)
		}
	}

}

func printError(err error) {
	e := err.Error()
	if strings.Split(e, ":")[0] == "rpc error" {
		es := strings.Split(e, "= ")

		em := es[2]
		errMsg := "Error: " + em

		ec := es[1]
		errCode := ec[0 : len(ec)-5]
		errCode = "Code: " + errCode

		fmt.Fprintln(os.Stderr, errCode, errMsg)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
