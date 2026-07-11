package logger

import (
	"io"
	"log/slog"
	"os"
)

func NewJSON(w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	return slog.New(slog.NewJSONHandler(w, nil))
}

func NewText(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	return slog.New(slog.NewTextHandler(w, opts))
}

func Default() *slog.Logger {
	return NewJSON(os.Stdout)
}
