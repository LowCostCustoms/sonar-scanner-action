package command

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Run runs a command and writes output from the process stdout and stderr
// into the specified logger entry.
func Run(entry *logrus.Entry, cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go redirectOutput(entry, logrus.InfoLevel, stdout)
	go redirectOutput(entry, logrus.ErrorLevel, stderr)

	return cmd.Wait()
}

func redirectOutput(entry *logrus.Entry, level logrus.Level, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		entry.Logln(level, scanner.Text())
	}
}
