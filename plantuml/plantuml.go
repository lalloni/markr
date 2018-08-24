package plantuml

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/lalloni/markr/logging"
	"github.com/lalloni/markr/processes"
	"github.com/mitchellh/go-homedir"
)

func downloadTo(target, url string) error {
	r, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("executing http get: %v", err)
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status %v", r.StatusCode)
	}
	rand.Seed(time.Now().UnixNano())
	tmp := path.Join(os.TempDir(), strconv.Itoa(rand.Int()))
	tmpf, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	defer func() {
		tmpf.Close()
		os.Remove(tmp)
	}()
	c, err := io.Copy(tmpf, r.Body)
	if err != nil {
		return fmt.Errorf("downloading: %v", err)
	}
	if c != r.ContentLength {
		return fmt.Errorf("downloaded %v bytes expecting %v", c, r.ContentLength)
	}
	if _, err := os.Stat(path.Dir(target)); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path.Dir(target), os.FileMode(0777))
			if err != nil {
				return fmt.Errorf("creating cache dir: %v", err)
			}
		} else {
			return fmt.Errorf("looking for cache dir: %v", err)
		}
	}
	ff, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("creating file: %v", err)
	}
	defer ff.Close()
	tmpf.Seek(0, 0)
	cc, err := io.Copy(ff, tmpf)
	if err != nil {
		return fmt.Errorf("copying to final path: %v", err)
	}
	if cc != c {
		return fmt.Errorf("copied %v bytes expecting %v", cc, c)
	}
	return nil
}

func Render(ctx context.Context, input io.Reader, output io.Writer, format string) error {
	log := logging.ZapLogger(ctx)
	p, err := homedir.Expand("~/.cache/markr/plantuml.jar")
	if err != nil {
		return fmt.Errorf("building plantuml.jar cached location: %v", err)
	}
	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			err := downloadTo(p, fmt.Sprintf("https://downloads.sourceforge.net/project/plantuml/plantuml.jar?r=https%%3A%%2F%%2Fsourceforge.net%%2Fprojects%%2Fplantuml%%2Ffiles%%2Fplantuml.jar%%2Fdownload%%3Fuse_mirror%%3Dautoselect&ts=%d&use_mirror=autoselect", time.Now().Unix()))
			if err != nil {
				return fmt.Errorf("downloading plantuml.jar to %q: %v", p, err)
			}
		} else {
			return fmt.Errorf("checking plantuml.jar cached existence: %v", err)
		}
	}
	log.Sugar().Infow("using plantuml from " + p)
	log.Sugar().Infow("rendering with plantuml", "format", format)
	logger := logging.LoggerWriter(log, "plantuml")
	defer logger.Close()
	cmd := exec.Command("java", "-jar", p, "-v", "-pipe", "-t"+format)
	err = processes.Pipe(ctx, cmd, input, output, logger)
	if err != nil {
		return fmt.Errorf("running plantuml: %v", err)
	}
	return nil
}
