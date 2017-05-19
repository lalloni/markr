package logging

import (
	"bufio"
	"context"
	"io"

	"go.uber.org/zap"
)

const (
	zapLogger = "zap.Logger"
)

func WithZapLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapLogger, logger)
}

func ZapLogger(ctx context.Context) *zap.Logger {
	return ctx.Value(zapLogger).(*zap.Logger)
}

func LoggerWriter(logger *zap.Logger, prefix string) io.WriteCloser {
	r, w := io.Pipe()
	s := bufio.NewScanner(r)
	go func() {
		for s.Scan() {
			logger.Info(prefix + ": " + s.Text())
		}
	}()
	return w
}
