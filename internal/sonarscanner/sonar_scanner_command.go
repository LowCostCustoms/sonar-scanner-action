package sonarscanner

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"

	"github.com/sirupsen/logrus"
)

var messagePrefixRegex = regexp.MustCompile("^(\\d+:\\d+:\\d+\\.\\d+\\s+)?(DEBUG|WARN|INFO|ERROR):\\s*")

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
		level, message := getLevelAndMessage(defaultLevel, scanner.Text())
		log.Logln(level, message)
	}
}

func getLevelAndMessage(defaultLevel logrus.Level, line string) (logrus.Level, string) {
	indices := messagePrefixRegex.FindStringSubmatchIndex(line)
	if indices != nil {
		message := line[indices[1]:]
		switch line[indices[4]:indices[5]] {
		case "DEBUG":
			return logrus.DebugLevel, message
		case "INFO":
			return logrus.InfoLevel, message
		case "WARN":
			return logrus.WarnLevel, message
		case "ERROR":
			return logrus.ErrorLevel, message
		}
	}

	return defaultLevel, line
}
