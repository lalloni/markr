package processes

import (
	"context"
	"fmt"
	"io"
	"os/exec"

	"go.uber.org/zap"

	"github.com/lalloni/markr/logging"
)

func Pipe(ctx context.Context, cmd *exec.Cmd, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {

	log := logging.ZapLogger(ctx)

	log.Info("running", zap.Strings("command", cmd.Args))

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("connecting stdin pipe: %v", err)
	}
	defer stdinPipe.Close()
	go func() {
		io.Copy(stdinPipe, stdin)
		stdinPipe.Close()
	}()

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("connecting stdout pipe: %v", err)
	}
	defer stdoutPipe.Close()
	go io.Copy(stdout, stdoutPipe)

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("connecting stderr pipe: %v", err)
	}
	defer stderrPipe.Close()
	go io.Copy(stderr, stderrPipe)

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("running command: %v", err)
	}

	return nil
}
