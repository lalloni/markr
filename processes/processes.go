package processes

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"go.uber.org/zap"
)

func RunCommand(cmd *exec.Cmd, log *zap.SugaredLogger) error {
	log.Infow("running", "command", cmd.Args)
	name := cmd.Args[0]
	o, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("connecting command %q stdout: %v", name, err)
	}
	defer o.Close()
	e, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("connecting command %q stderr: %v", name, err)
	}
	defer e.Close()
	s := bufio.NewScanner(io.MultiReader(o, e))
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("starting command %q: %v", name, err)
	}
	for s.Scan() {
		log.Infof("command %q output: %s", name, s.Text())
	}
	if s.Err() != nil {
		log.Errorf("reading command %q output: %v", name, s.Err())
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("waiting command %q exit: %v", name, err)
	}
	return nil
}
