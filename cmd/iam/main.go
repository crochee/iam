package main

import (
	"log/slog"
	"os"

	"github.com/crochee/iam/internal"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("hello world!", slog.String("version", internal.Version))
}
