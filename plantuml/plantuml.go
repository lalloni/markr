package plantuml

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"go.uber.org/zap"

	"github.com/lalloni/markr/processes"
)

func Render(input string, format string, log *zap.SugaredLogger) (string, error) {
	log.Infow("rendering with plantuml", "input", input, "format", format)
	cmd := exec.Command("plantuml", "-v", "-t"+format, input)
	err := processes.RunCommand(cmd, log)
	if err != nil {
		return "", fmt.Errorf("running plantuml: %v", err)
	}
	output := strings.TrimSuffix(input, path.Ext(input)) + "." + format
	if _, err := os.Stat(output); os.IsNotExist(err) {
		return "", fmt.Errorf("image file %q not created", output)
	}
	return output, nil
}
