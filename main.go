package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/lalloni/markr/inkscape"
	"github.com/lalloni/markr/pandoc"
	"github.com/lalloni/markr/plantuml"

	"go.uber.org/zap"
)

const (
	MacroBeginDelimiter = "{{"
	MacroEndDelimiter   = "}}"
)

var (
	inputFile  = flag.String("in", "FILE", "Markdown input file")
	outputFile = flag.String("out", "FILE", "PDF output file")
	keep       = flag.Bool("keep", false, "Keep temporal files")
	logger     *zap.Logger
	log        *zap.SugaredLogger
	err        error
)

func init() {
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

var deletes []string

func onExitDeletes() {
	for _, file := range deletes {
		log.Infow("execute delete on exit", "file", file)
		delete(file)
	}
}

func deleteOnExit(file string) {
	log.Infow("delete on exit", "file", file)
	deletes = append(deletes, file)
}

func delete(file string) error {
	if *keep {
		log.Infow("keeping", "file", file)
		return nil
	}
	log.Infow("deleting", "file", file)
	return os.Remove(file)
}

func tempFile(kind string) (*os.File, error) {
	file, err := ioutil.TempFile(os.TempDir(), "markp-"+kind+"-")
	deleteOnExit(file.Name())
	return file, err
}

func changeExtension(file string, ext string) string {
	return strings.TrimSuffix(file, path.Ext(file)) + "." + ext
}

func isMacroStart(line string) bool {
	return strings.HasPrefix(line, MacroBeginDelimiter)
}

func isMacroEnd(line string) bool {
	return strings.HasPrefix(line, MacroEndDelimiter)
}

func isPlantUMLMacroStart(line string) bool {
	return isMacroStart(line) && strings.HasPrefix(strings.TrimPrefix(line, MacroBeginDelimiter), "plantuml")
}

func main() {
	flag.Parse()

	logger = logger.WithOptions()

	defer logger.Sync()
	defer onExitDeletes()

	inf, err := os.Open(*inputFile)
	if err != nil {
		log.Errorw("opening file for reading", "file", *inputFile)
		return
	}
	defer inf.Close()

	mdf, err := tempFile("markdown")
	if err != nil {
		log.Errorw("creating temporary markdown file", "error", err)
		return
	}
	defer mdf.Close()
	mdw := bufio.NewWriter(mdf)

	var umlf *os.File
	var umlw *bufio.Writer
	var line, lineprev string

	macro := false
	ins := bufio.NewScanner(bufio.NewReader(inf))
	for ins.Scan() {
		lineprev = line
		line = ins.Text()
		log.Infow("read", "line", line)
		if macro {
			if isMacroEnd(line) {
				log.Info("macro end")
				if lineprev[0:1] != "@" {
					log.Info("generating @enduml marker")
					_, err := umlw.WriteString("@enduml\n")
					if err != nil {
						log.Errorw("writing to temporary file", "error", err)
						return
					}
				}
				err = umlw.Flush()
				if err != nil {
					log.Errorw("flushing temporary file", "error", err)
					return
				}
				err = umlf.Close()
				if err != nil {
					log.Errorw("closing temporary file", "error", err)
					return
				}
				svg, err := plantuml.Render(umlf.Name(), "svg", log)
				if err != nil {
					log.Errorw("rendering with plantuml", "error", err, "source", umlf.Name())
					return
				}
				deleteOnExit(svg)
				pdf := changeExtension(svg, "pdf")
				err = inkscape.Convert(svg, pdf, log)
				if err != nil {
					log.Errorw("converting with inkscape", "error", err, "source", svg)
					return
				}
				deleteOnExit(pdf)
				_, err = fmt.Fprintf(mdw, "![](%s)\\ \n", pdf)
				if err != nil {
					log.Errorw("writing to output", "error", err)
					return
				}
				umlw = nil
				umlf = nil
				macro = false
			} else {
				if isPlantUMLMacroStart(lineprev) && line[0:1] != "@" {
					log.Info("generating @startuml marker")
					_, err = umlw.WriteString("@startuml\n")
					if err != nil {
						log.Errorw("writing to file", "error", err)
						return
					}
				}
				log.Infow("copying line to file", "file", umlf.Name())
				_, err = umlw.WriteString(line + "\n")
				if err != nil {
					log.Errorw("writing to file", "file", umlf.Name(), "error", err)
					return
				}
			}
		} else { // not in macro
			if isPlantUMLMacroStart(line) {
				macro = true
				umlf, err = tempFile("plantuml")
				if err != nil {
					log.Errorw("creating temporary file", "error", err)
					return
				}
				log.Infow("macro start", "file", umlf.Name())
				umlw = bufio.NewWriter(umlf)
				continue
			}
			log.Info("copying line to output")
			_, err = mdw.WriteString(line + "\n")
			if err != nil {
				log.Errorw("writing to output", "error", err)
				return
			}
		}
	}
	if ins.Err() != nil {
		log.Errorw("reading input file", "error", ins.Err())
	}
	if macro {
		umlw.Flush()
		umlf.Close()
	}
	mdw.Flush()
	mdf.Close()
	err = pandoc.RenderMarkdown(mdf.Name(), *outputFile, log)
	if err != nil {
		log.Errorw("rendering markdown", "error", err)
		return
	}
	log.Infow("generated pdf file", "file", *outputFile)
}
