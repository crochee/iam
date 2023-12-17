package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crochee/iam/internal"
	"github.com/crochee/iam/internal/router"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/multierr"
)

var configFile = flag.String("f", "./config/template.yaml", "the config file")

func main() {
	flag.Parse()
	if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.InfoContext(ctx, "hello world!", slog.String("version", internal.Version))
	// gin 设置
	// gin.DefaultWriter = writer
	// 初始化数据库

	g := pool.New().WithContext(context.Background()).WithCancelOnError()
	srv := &http.Server{
		Addr:    ":31000",
		Handler: router.New(),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	slog.DebugContext(ctx, "listen on %s", srv.Addr)
	// 服务启动流程
	g.Go(func(ctx context.Context) error {
		return startAction(ctx, srv)
	})
	// 服务关闭流程
	g.Go(func(ctx context.Context) error {
		return shutdownAction(ctx, srv)
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func startAction(ctx context.Context, srv *http.Server) error {
	return srv.ListenAndServe()
}

const DefaultStopTime = 15 * time.Second

func shutdownAction(ctx context.Context, srv *http.Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	var err error
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case <-quit:
	}
	newCtx, cancel := context.WithTimeout(ctx, DefaultStopTime)
	defer cancel()
	return multierr.Append(err, srv.Shutdown(newCtx))
}
