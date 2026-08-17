package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func newSignalContext(parent context.Context, _ func()) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
}
