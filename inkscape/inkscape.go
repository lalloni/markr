package inkscape

import (
	"fmt"
	"os/exec"

	"go.uber.org/zap"

	"github.com/lalloni/markr/processes"
)

func Convert(svg string, pdf string, log *zap.SugaredLogger) error {
	log.Infow("rendering with inkscape", "input", svg, "output", pdf)
	cmd := exec.Command("inkscape", "--without-gui", "--export-dpi", "300", "--file", svg, "--export-pdf", pdf)
	err := processes.RunCommand(cmd, log)
	if err != nil {
		return fmt.Errorf("running inkscape: %v", err)
	}
	return nil
}
