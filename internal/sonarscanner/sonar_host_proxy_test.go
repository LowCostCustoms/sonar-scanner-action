package sonarscanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSonarHostProxyNew(t *testing.T) {
	factory := &sonarHostProxyFactory{
		listenAddr:   "localhost:9999",
		sonarHostUrl: "http://sonarqube.local",
	}

	proxy, err := factory.new()

	assert.NotNil(t, proxy)
	assert.Nil(t, err)
}

func TestSonarHostProxyNewInvalidUrl(t *testing.T) {
	factory := &sonarHostProxyFactory{
		listenAddr:   "localhost:9999",
		sonarHostUrl: "\x00\x01",
	}

	proxy, err := factory.new()

	assert.NotNil(t, err)
	assert.Nil(t, proxy)
}
