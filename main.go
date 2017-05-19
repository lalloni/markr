package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/lalloni/markr/fileutils"
	"github.com/lalloni/markr/inkscape"
	"github.com/lalloni/markr/logging"
	"github.com/lalloni/markr/options"
	"github.com/lalloni/markr/pandoc"
	"github.com/lalloni/markr/plantuml"
)

const (
	BeginDelimiter = "{{plantuml"
	EndDelimiter   = "}}"
)

func isMacroStart(line string) bool {
	return strings.HasPrefix(line, BeginDelimiter)
}

func isMacroEnd(line string) bool {
	return strings.HasPrefix(line, EndDelimiter)
}

func generatePDF(ctx context.Context, uml io.Reader, diagram string) error {
	var svg bytes.Buffer
	err := plantuml.Render(ctx, uml, &svg, "svg")
	if err != nil {
		return fmt.Errorf("rendering with plantuml: %v", err)
	}
	var pdf bytes.Buffer
	err = inkscape.ConvertToPDF(ctx, &svg, &pdf)
	if err != nil {
		return fmt.Errorf("converting with inkscape: %v", err)
	}
	err = ioutil.WriteFile(diagram, pdf.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("writing PDF file: %v", err)
	}
	return nil
}

func generateEPS(ctx context.Context, uml io.Reader, diagram string) error {
	var eps bytes.Buffer
	err := plantuml.Render(ctx, uml, &eps, "eps")
	if err != nil {
		return fmt.Errorf("rendering with plantuml: %v", err)
	}
	err = ioutil.WriteFile(diagram, eps.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("writing PDF file: %v", err)
	}
	return nil
}

func main() {
	options.ConfigureFlags()
	flag.Parse()

	ctx := context.Background()

	ctx = options.WithOptions(ctx)
	opts := options.Get(ctx)

	if opts.Usage {
		flag.Usage()
		return
	}

	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeCaller = nil
	if !opts.Verbose {
		loggerConfig.Level.SetLevel(zap.ErrorLevel)
	}
	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	log := logger.Sugar()
	ctx = logging.WithZapLogger(ctx, logger)

	defer fileutils.DoDeletes(ctx)

	inf, err := os.Open(opts.InputFile)
	if err != nil {
		log.Errorw("opening file for reading", "file", opts.InputFile)
		return
	}
	defer inf.Close()

	var markdown, uml bytes.Buffer
	var base, line, lineprev string
	var fixingPlantUML bool

	{
		sha1 := sha1.Sum([]byte(opts.InputFile))
		base = "markr-" + hex.EncodeToString(sha1[:])
	}

	macro := false
	ins := bufio.NewScanner(bufio.NewReader(inf))
	for ins.Scan() {
		lineprev = line
		line = ins.Text()
		log.Infow("read", "line", line)
		if macro {

			if isMacroEnd(line) {

				log.Info("macro end")

				if fixingPlantUML {
					log.Info("generating @enduml marker")
					uml.WriteString("@enduml\n")
				}

				sha1 := sha1.Sum(uml.Bytes())
				sha1hex := hex.EncodeToString(sha1[:])
				log.Infow("uml source checksum", "sha1", sha1hex)

				diagram := fileutils.NewTempFileName(base, "diagram", sha1hex, opts.Diagrams)

				if _, err = os.Stat(diagram); opts.Cache && os.IsNotExist(err) || !opts.Cache {

					switch opts.Diagrams {
					case "pdf":
						err = generatePDF(ctx, &uml, diagram)
					case "eps":
						err = generateEPS(ctx, &uml, diagram)
					default:
						err = fmt.Errorf("unknown diagram format: %q", opts.Diagrams)
					}

					if err != nil {
						log.Errorw("generating diagram", "error", err, "source", uml.String())
						return
					}

					if !opts.Cache {
						fileutils.AddDelete(ctx, diagram)
					} else {
						log.Infow("keeping for cache", "file", diagram)
					}

				}

				log.Infow("using diagram", "file", diagram)
				_, err = fmt.Fprintf(&markdown, "![](%s)\n", diagram)
				if err != nil {
					log.Errorw("writing to output", "error", err)
					return
				}

				uml.Reset()
				macro = false
				fixingPlantUML = false

			} else {

				if isMacroStart(lineprev) && line[0:1] != "@" {
					log.Info("generating @startuml marker")
					fixingPlantUML = true
					uml.WriteString("@startuml\n")
				}
				log.Info("keeping plantuml line")
				uml.Write([]byte(line))
				uml.Write([]byte("\n"))

			}

		} else {

			// not in macro

			if isMacroStart(line) {
				log.Info("starting plantuml macro")
				macro = true
				continue
			}

			log.Info("copying line")
			_, err = markdown.WriteString(line + "\n")
			if err != nil {
				log.Errorw("writing to output", "error", err)
				return
			}

		}
	}

	if ins.Err() != nil {
		log.Errorw("reading input file", "error", ins.Err())
		return
	}

	err = pandoc.RenderMarkdown(ctx, &markdown, opts.OutputFile)
	if err != nil {
		log.Errorw("rendering markdown", "error", err)
		return
	}

}
