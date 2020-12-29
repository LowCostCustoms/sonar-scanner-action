package environment

import (
	"fmt"
	"reflect"
	"time"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

type Environment struct {
	SonarHostUrl           string        `env:"SONAR_HOST_URL"`
	SonarHostCert          string        `env:"SONAR_HOST_CERT"`
	ProjectFileLocation    string        `env:"PROJECT_FILE_LOCATION" envDefault:""`
	WaitForQualityGate     bool          `env:"WAIT_FOR_QUALITY_GATE" envDefault:"true"`
	QualityGateWaitTimeout time.Duration `env:"QUALITY_GATE_WAIT_TIMEOUT" envDefault:"2m"`
	LogLevel               logrus.Level  `env:"LOG_LEVEL" envDefault:"info"`
	TlsSkipVerify          bool          `env:"TLS_SKIP_VERIFY" envDefault:"false"`
}

func Get() (*Environment, error) {
	parsers := env.CustomParsers{
		reflect.TypeOf(logrus.DebugLevel): func(str string) (interface{}, error) {
			return logrus.ParseLevel(str)
		},
	}

	environment := &Environment{}
	if err := env.ParseWithFuncs(environment, parsers); err != nil {
		return nil, err
	}

	if environment.QualityGateWaitTimeout <= time.Duration(0) {
		return nil, fmt.Errorf("quality gate wait timeout must be positive")
	}

	return environment, nil
}
