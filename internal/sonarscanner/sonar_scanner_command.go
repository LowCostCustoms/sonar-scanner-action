package sonarscanner

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func runSonarScanner(log *logrus.Entry, cmd *exec.Cmd) error {
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

	go redirectOutput(log, logrus.InfoLevel, stdout)
	go redirectOutput(log, logrus.WarnLevel, stderr)

	return cmd.Wait()
}

func redirectOutput(log *logrus.Entry, defaultLevel logrus.Level, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Logln(defaultLevel, scanner.Text())
	}
}
