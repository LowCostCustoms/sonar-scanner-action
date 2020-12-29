package command

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Run(log *logrus.Entry, command *exec.Cmd) error {
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	go redirectOutput(stdout, log, logrus.InfoLevel)
	go redirectOutput(stderr, log, logrus.WarnLevel)

	return command.Wait()
}

func redirectOutput(output io.Reader, entry *logrus.Entry, level logrus.Level) {
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		entry.Logln(level, scanner.Text())
	}
}
