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

const exitError = 1

func main() {
	if err := run(); err != nil {
		os.Exit(exitError)
	}
}

func run() error {
	cliConfig, err := cli.LoadConfig()
	if err != nil {
		fmt.Println(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if cmd, err := cli.New(cliConfig).ExecuteContextC(ctx); err != nil {
		printError(err)

		switch errStr := err.Error(); {
		case strings.HasPrefix(errStr, "unknown command"),
			strings.HasPrefix(errStr, "unknown flag"),
			strings.HasPrefix(errStr, "unknown shorthand flag"):
			if !strings.HasSuffix(errStr, "\n") {
				fmt.Println()
			}
			fmt.Println(cmd.UsageString())
			return nil
		}

		return err
	}

	return nil
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
