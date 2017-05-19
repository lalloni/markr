package pandoc

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/lalloni/markr/logging"
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
	"-f", "markdown-implicit_figures",
	"-t", "latex",
	"-V", "mainfont=Ubuntu",
	"-V", "monofont=Iosevka",
	"-V", "papersize=A4",
	"-V", "urlcolor=blue",
	"-V", "colorlinks",
	"-V", "toccolor=blue",
	"-V", "geometry=margin=2cm",
	"-V", "lang=spanish",
	"-V", "babel-lang=spanish",
	"-V", "include-before=\\addto\\captionsspanish{\\renewcommand{\\contentsname}{Contenidos}}",
	"-V", "include-before=\\renewcommand{\\contentsname}{Contenidos}",
}

func RenderMarkdown(ctx context.Context, input io.Reader, file string) error {
	log := logging.ZapLogger(ctx)
	log.Info("rendering with pandoc")
	cmd := exec.Command("pandoc", append(pandocOptions, "-o", file)...)
	logger := logging.LoggerWriter(log, "pandoc")
	defer logger.Close()
	err := processes.Pipe(ctx, cmd, input, logger, logger)
	if err != nil {
		return fmt.Errorf("running pandoc: %v", err)
	}
	return nil
}
