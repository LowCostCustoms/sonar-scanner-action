package environment

import (
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	setEnvironment()

	e, err := Get()

	assert.Nil(t, err)
	assert.Equal(t, e.SonarHostCert, "sonar-host-cert")
	assert.Equal(t, e.SonarHostUrl, "sonar-host-url")
	assert.Equal(t, e.ProjectFileLocation, "project-file-location")
	assert.Equal(t, e.WaitForQualityGate, true)
	assert.Equal(t, e.QualityGateWaitTimeout, 10*time.Second)
	assert.Equal(t, e.LogLevel, logrus.WarnLevel)
	assert.Equal(t, e.SonarLogin, "sonar-login")
	assert.Equal(t, e.SonarPassword, "sonar-password")
}

func TestGetParseFailed(t *testing.T) {
	setEnvironment()

	os.Setenv("WAIT_FOR_QUALITY_GATE", "yikes")

	e, err := Get()

	assert.NotNil(t, err)
	assert.Nil(t, e)
}

func TestGetInvalidDuration(t *testing.T) {
	setEnvironment()

	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "-10s")

	e, err := Get()

	assert.NotNil(t, err)
	assert.Nil(t, e)
}

func setEnvironment() {
	os.Setenv("SONAR_HOST_URL", "sonar-host-url")
	os.Setenv("SONAR_HOST_CERT", "sonar-host-cert")
	os.Setenv("PROJECT_FILE_LOCATION", "project-file-location")
	os.Setenv("WAIT_FOR_QUALITY_GATE", "true")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "10s")
	os.Setenv("LOG_LEVEL", "warning")
	os.Setenv("SONAR_LOGIN", "sonar-login")
	os.Setenv("SONAR_PASSWORD", "sonar-password")
}
