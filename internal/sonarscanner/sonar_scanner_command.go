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
	rx, _ := regexp.Compile("^\\d+:\\d+:\\d+\\.\\d+\\s+(DEBUG|INFO|WARNING|ERROR)?:\\s*(.*)?$")

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		matches := rx.FindStringSubmatch(scanner.Text())
		if matches == nil {
			log.Logln(defaultLevel, scanner.Text())
		} else {
			message := matches[2]
			switch matches[1] {
			case "DEBUG":
				log.Debug(message)
				break
			case "INFO":
				log.Info(message)
				break
			case "WARNING":
				log.Warn(message)
				break
			case "ERROR":
				log.Error(message)
				break
			default:
				log.Logln(defaultLevel, message)
				break
			}
		}
	}
}
