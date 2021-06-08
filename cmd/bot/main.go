package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
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

	// dont need log rotation yet
	//w := zapcore.AddSync(&lumberjack.Logger{
	//	Filename:   "./bot.log",
	//	MaxSize:    500, // megabytes
	//	MaxBackups: 3,
	//	MaxAge:     28, // days
	//})

	//core := zapcore.NewCore(
	//	zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
	//	zapcore.NewMultiWriteSyncer(w, zapcore.AddSync(os.Stdout)),
	//	zap.InfoLevel,
	//)
	//logger := zap.New(core)

	conf := zap.NewProductionConfig()
	conf.OutputPaths = append(conf.OutputPaths, logFile)
	logger, err := conf.Build()
	if err != nil {
		return xerrors.Errorf("can't init logger: %w", err)
	}

	defer func() { _ = logger.Sync() }()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return runBot(ctx, logger.Named("bot"), logFile)
	})
	g.Go(func() error {
		// no metrics yet
		//return runMetrics(ctx, logger.Named("metrics"))
		return nil
	})

	return g.Wait()
}
