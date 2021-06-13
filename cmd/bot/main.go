package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/xerrors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// listen for 3 signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	// for local development
	if err := loadEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
	}

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}

func run(ctx context.Context) error {
	logFile := fmt.Sprintf("bot-%s.log", time.Now().Format("2006-01-02T15-04-05"))

	conf := zap.NewProductionConfig()
	conf.OutputPaths = append(conf.OutputPaths, logFile)
	logger, err := conf.Build()
	if err != nil {
		return xerrors.Errorf("can't init logger: %w", err)
	}

	defer func() { _ = logger.Sync() }()

	return runBot(ctx, logger.Named("bot"), logFile)
}
