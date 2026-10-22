package main

import (
	"context"
	"os"
	"os/signal"
)

func signalNotifyContext(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt)
}
