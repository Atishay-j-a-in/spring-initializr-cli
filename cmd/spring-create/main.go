package main

import (
	"context"
	"fmt"
	"os"

	"example.com/spring-create/internal/app"
)

// version is replaced with the v-prefixed tag version by GoReleaser.
var version = "dev"

func main() {
	ctx, stop := signalContext()
	defer stop()

	if err := app.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version); err != nil {
		if code := app.ExitCode(err); code != app.ExitSuccess {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(code)
		}
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	return signalNotifyContext(context.Background())
}
