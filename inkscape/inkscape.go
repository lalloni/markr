package inkscape

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/lalloni/markr/logging"
	"github.com/lalloni/markr/options"
	"github.com/lalloni/markr/processes"
)

func ConvertToPDF(ctx context.Context, input io.Reader, output io.Writer) error {
	log := logging.ZapLogger(ctx)
	log.Info("rendering with inkscape")
	logger := logging.LoggerWriter(log, "inkscape")
	defer logger.Close()
	cmd := exec.Command("inkscape", "--without-gui", "--export-dpi", strconv.Itoa(options.Get(ctx).DiagramsDPI), "--file", "/dev/stdin", "--export-pdf", "/dev/stdout")
	err := processes.Pipe(ctx, cmd, input, output, logger)
	if err != nil {
		return fmt.Errorf("rendering with inkscape: %v", err)
	}
	return nil
}
