package config

import (
	"fmt"
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
)

// Environment is a structure containing action environment variables.
type Environment struct {
	SonarHostURL           string
	SonarHostCertificate   string
	ProjectFileLocation    string
	WaitForQualityGate     bool
	QualityGateWaitTimeout time.Duration
}

// GetEnvironment returns an Environment structure containing
func GetEnvironment() (*Environment, error) {
	environment := &Environment{
		SonarHostURL:         os.Getenv(sonarHostURLVarName),
		SonarHostCertificate: os.Getenv(sonarHostCertificateVarName),
		ProjectFileLocation:  os.Getenv(projectFileLocationVarName),
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

		if timeout < 0 {
			return nil, fmt.Errorf(
				"%s may not be negative",
				qualityGateWaitTimeoutVarName,
			)
		}

		environment.QualityGateWaitTimeout = timeout
	}

	return environment, nil
}
