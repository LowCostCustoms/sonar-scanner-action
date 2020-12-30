package sonarscanner

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadProjectProperties(t *testing.T) {
	tempDir := t.TempDir()
	propertiesFileName := path.Join(tempDir, "properties.properties")

	file, _ := os.Create(propertiesFileName)
	file.WriteString(`
    sonar.host.url = http://sonarqube.local
    sonar.login = sonar-login with whitespace
    sonar.password = sonar@passw@0rd
    `)
	file.Close()

	props, err := readProjectProperties(propertiesFileName)

	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, props.sonarHostUrl, "http://sonarqube.local")
	assert.Equal(t, props.login, "sonar-login with whitespace")
	assert.Equal(t, props.password, "sonar@passw@0rd")
}

func TestReadProjectPropertiesInvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	propertiesFileName := path.Join(tempDir, "non-existent.properties")

	props, err := readProjectProperties(propertiesFileName)

	assert.Nil(t, props)
	assert.NotNil(t, err)
}
