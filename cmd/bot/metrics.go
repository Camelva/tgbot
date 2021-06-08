package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// runMetrics unneeded yet
func _(ctx context.Context, logger *zap.Logger) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	attachProfiler(mux)
	server := &http.Server{Addr: fmt.Sprintf(":%s", port), Handler: mux}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		logger.Info("ListenAndServe", zap.String("addr", server.Addr))
		return server.ListenAndServe()
	})
	grp.Go(func() error {
		<-ctx.Done()

		shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		logger.Info("Shutdown", zap.String("addr", server.Addr))
		if err := server.Shutdown(shutCtx); err != nil {
			return multierr.Append(err, server.Close())
		}
		return nil
	})

	return grp.Wait()
}

func attachProfiler(router *http.ServeMux) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
}
