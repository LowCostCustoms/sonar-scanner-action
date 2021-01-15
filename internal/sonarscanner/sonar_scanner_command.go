package sonarscanner

import (
	"bufio"
	"io"
	"os/exec"
	"regexp"

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
	rx := regexp.MustCompile("^(INFO|WARN|ERRO|DEBU)")

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		matches := rx.FindStringSubmatch(line)

		level := defaultLevel
		if matches != nil {
			switch matches[1] {
			case "WARN":
				level = logrus.WarnLevel
				break
			case "INFO":
				level = logrus.InfoLevel
				break
			case "DEBU":
				level = logrus.DebugLevel
				break
			case "ERRO":
				level = logrus.ErrorLevel
				break
			}
		}
		log.Logln(level, line)
	}
}
