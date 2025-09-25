package logger

import (
	"log/slog"
	"os"
	"sync"

	"github.com/lmittmann/tint"
)

var once sync.Once
var instance *slog.Logger

func Get() *slog.Logger {
	once.Do(func() {
		handler := tint.NewHandler(os.Stdout, &tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
		})

		instance = slog.New(handler)
	})

	return instance
}
