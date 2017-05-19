package plantuml

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/lalloni/markr/logging"
	"github.com/lalloni/markr/processes"
)

func Render(ctx context.Context, input io.Reader, output io.Writer, format string) error {
	log := logging.ZapLogger(ctx)
	log.Sugar().Infow("rendering with plantuml", "format", format)
	logger := logging.LoggerWriter(log, "plantuml")
	defer logger.Close()
	cmd := exec.Command("plantuml", "-v", "-pipe", "-t"+format)
	err := processes.Pipe(ctx, cmd, input, output, logger)
	if err != nil {
		return fmt.Errorf("running plantuml: %v", err)
	}
	return nil
}
