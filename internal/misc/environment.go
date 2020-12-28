package misc

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const (
	sonarHostURLVarName           = "SONAR_HOST_URL"
	sonarHostCertificateVarName   = "SONAR_HOST_CERTIFICATE"
	projectFileLocationVarName    = "PROJECT_FILE_LOCATION"
	waitForQualityGateVarName     = "WAIT_FOR_QUALITY_GATE"
	qualityGateWaitTimeoutVarName = "QUALITY_GATE_WAIT_TIMEOUT"
	logLevelVarName               = "LOG_LEVEL"
)

type Environment struct {
	SonarHostURL           string
	SonarHostCert          string
	ProjectFileLocation    string
	WaitForQualityGate     bool
	LogLevel               logrus.Level
	QualityGateWaitTimeout time.Duration
}

func GetEnvironment() (*Environment, error) {
	environment := &Environment{
		SonarHostURL:        os.Getenv(sonarHostURLVarName),
		SonarHostCert:       os.Getenv(sonarHostCertificateVarName),
		ProjectFileLocation: os.Getenv(projectFileLocationVarName),
		LogLevel:            logrus.InfoLevel,
	}

	waitForQualityGate, err := strconv.ParseBool(
		os.Getenv(waitForQualityGateVarName),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse boolean from %s: %s",
			waitForQualityGateVarName,
			err,
		)
	}

	environment.WaitForQualityGate = waitForQualityGate
	if waitForQualityGate {
		timeout, err := time.ParseDuration(
			os.Getenv(qualityGateWaitTimeoutVarName),
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to parse duration from %s: %s",
				qualityGateWaitTimeoutVarName,
				err,
			)
		}

		if timeout <= 0 {
			return nil, fmt.Errorf(
				"%s may not be negative",
				qualityGateWaitTimeoutVarName,
			)
		}

		environment.QualityGateWaitTimeout = timeout
	}

	logLevelString := os.Getenv(logLevelVarName)
	if logLevelString != "" {
		logLevel, err := logrus.ParseLevel(logLevelString)
		if err != nil {
			return nil, err
		}

		environment.LogLevel = logLevel
	}

	return environment, nil
}
