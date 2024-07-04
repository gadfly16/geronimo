package main

import (
	"log/slog"
	"runtime"

	"github.com/gadfly16/geronimo/cli"
)

func main() {
	cli.Execute()
	slog.Info("Goroutinge count on exit.", "numGoroutine", runtime.NumGoroutine())
}
