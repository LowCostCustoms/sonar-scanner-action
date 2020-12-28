package misc

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGetEnvironment(t *testing.T) {
	os.Setenv("SONAR_HOST_URL", "sonar-host")
	os.Setenv("SONAR_HOST_CERTIFICATE", "sonar-host-certificate")
	os.Setenv("PROJECT_FILE_LOCATION", "project-file-location")
	os.Setenv("WAIT_FOR_QUALITY_GATE", "0")

	env, err := GetEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, env.SonarHostURL, "sonar-host")
	assert.Equal(t, env.SonarHostCert, "sonar-host-certificate")
	assert.Equal(t, env.ProjectFileLocation, "project-file-location")
	assert.Equal(t, env.WaitForQualityGate, false)
	assert.Equal(t, env.QualityGateWaitTimeout, time.Duration(0))
}

func TestGetEnvironmentWithWaitTimeout(t *testing.T) {
	os.Setenv("WAIT_FOR_QUALITY_GATE", "1")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "2s")

	env, err := GetEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, env.WaitForQualityGate, true)
	assert.Equal(t, env.QualityGateWaitTimeout, 2*time.Second)
}

func TestGetEnvironmentWithWrongWaitTimeout(t *testing.T) {
	os.Setenv("WAIT_FOR_QUALITY_GATE", "1")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "two seconds")

	_, err := GetEnvironment()

	assert.NotNil(t, err)
}

func TestGetEnvironmentWithWrongWaitForQualityGate(t *testing.T) {
	os.Setenv("WAIT_FOR_QUALITY_GATE", "nope")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "2s")

	_, err := GetEnvironment()

	assert.NotNil(t, err)
}
