package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	seconds, err := parseDuration("10s")

	assert.Nil(t, err)
	assert.Equal(t, seconds, time.Second*10)

	minutes, err := parseDuration("20m")

	assert.Nil(t, err)
	assert.Equal(t, minutes, time.Minute*20)

	hours, err := parseDuration("30h")

	assert.Nil(t, err)
	assert.Equal(t, hours, time.Hour*30)
}

func TestParseDurationInvalidFormat(t *testing.T) {
	_, err := parseDuration("30")

	assert.NotNil(t, err)

	_, err = parseDuration("30min")

	assert.NotNil(t, err)

	_, err = parseDuration("-10s")

	assert.NotNil(t, err)
}

func TestGetEnvironment(t *testing.T) {
	os.Setenv("PROJECT_FILE_LOCATION", "location")
	os.Setenv("WAIT_FOR_QUALITY_GATE", "false")
	os.Setenv("SONAR_HOST_URL", "host-url")
	os.Setenv("SONAR_SERVER_CERTIFICATE", "cert")

	env, err := GetEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, env.ProjectFileLocation, "location")
	assert.Equal(t, env.WaitForQualityGate, false)
	assert.Equal(t, env.WaitTimeout, time.Duration(0))
	assert.Equal(t, env.SonarHostURL, "host-url")
	assert.Equal(t, env.SonarServerCertificate, "cert")

	os.Setenv("WAIT_FOR_QUALITY_GATE", "true")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "5s")

	env, err = GetEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, env.WaitForQualityGate, true)
	assert.Equal(t, env.WaitTimeout, time.Second*time.Duration(5))

	os.Setenv("PROJECT_FILE_LOCATION", "")

	env, err = GetEnvironment()

	assert.Nil(t, err)
	assert.Equal(t, env.ProjectFileLocation, "sonar-project.properties")
}

func TestGetEnvironmentPropagatesError(t *testing.T) {
	os.Setenv("WAIT_FOR_QUALITY_GATE", "bruh")

	env, err := GetEnvironment()

	assert.Nil(t, env)
	assert.NotNil(t, err)

	os.Setenv("WAIT_FOR_QUALITY_GATE", "true")
	os.Setenv("QUALITY_GATE_WAIT_TIMEOUT", "-1")

	env, err = GetEnvironment()

	assert.Nil(t, env)
	assert.NotNil(t, err)
}

func TestReadSonarProjectProperties(t *testing.T) {
	tempFile, _ := ioutil.TempFile("", "scanner-config")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

}
