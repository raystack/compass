package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"connectrpc.com/connect"
	"github.com/raystack/compass/cli"
	"github.com/raystack/compass/internal/config"
	saltconfig "github.com/raystack/salt/config"
)

const (
	exitOK    = 0
	exitError = 1
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if cmd, err := cli.New(cfg).ExecuteContextC(ctx); err != nil {
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

func loadConfig() (*config.Config, error) {
	var cfg config.Config

	err := saltconfig.NewLoader(
		saltconfig.WithFile("./config.yaml"),
		saltconfig.WithEnvPrefix("COMPASS"),
	).Load(&cfg)
	if err != nil {
		loader := saltconfig.NewLoader(saltconfig.WithAppConfig("compass"))
		if loadErr := loader.Load(&cfg); loadErr != nil {
			return &cfg, fmt.Errorf("config not found: run \"compass config init\" to create one")
		}
	}
	return &cfg, nil
}

func printError(err error) {
	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		fmt.Fprintf(os.Stderr, "Code: %s Error: %s\n", connectErr.Code(), connectErr.Message())
		return
	}
	fmt.Fprintln(os.Stderr, err)
}
