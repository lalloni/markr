package pandoc

import (
	"fmt"
	"os"
	"os/exec"

	"go.uber.org/zap"

	"github.com/lalloni/markr/processes"
)

var pandocOptions = []string{
	"--toc",
	"--reference-links",
	"--smart",
	"--latex-engine=xelatex",
	"--number-sections",
	"--section-divs",
	"--standalone",
	"--self-contained",
	"-V", "mainfont=Ubuntu",
	"-V", "monofont=Iosevka",
	"-V", "papersize=A4",
	"-V", "urlcolor=blue",
	"-V", "colorlinks",
	"-V", "toccolor=blue",
	"-V", "geometry=margin=2cm",
	"-f", "markdown",
}

func RenderMarkdown(input string, output string, log *zap.SugaredLogger) error {
	log.Infow("rendering with pandoc", "input", input, "output", output)
	a := append(pandocOptions, "-o", output, input)
	cmd := exec.Command("pandoc", a...)
	err := processes.RunCommand(cmd, log)
	if err != nil {
		return fmt.Errorf("running pandoc: %v", err)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		return fmt.Errorf("output file %q not created", output)
	}
	return nil
}
