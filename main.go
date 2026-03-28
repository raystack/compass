package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/raystack/compass/cli"
	"google.golang.org/grpc/status"
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

		msg := err.Error()
		if strings.HasPrefix(msg, "unknown command") ||
			strings.HasPrefix(msg, "unknown flag") ||
			strings.HasPrefix(msg, "unknown shorthand flag") {
			if !strings.HasSuffix(msg, "\n") {
				fmt.Println()
			}
			fmt.Println(cmd.UsageString())
			os.Exit(exitOK)
		}
		os.Exit(exitError)
	}
}

func printError(err error) {
	if s, ok := status.FromError(err); ok {
		fmt.Fprintf(os.Stderr, "Code: %s Error: %s\n", s.Code(), s.Message())
		return
	}
	fmt.Fprintln(os.Stderr, err)
}
