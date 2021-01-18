package sonarscanner

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetLevelAndMessage(t *testing.T) {
	assertLevelAndMessage(
		t,
		logrus.InfoLevel,
		"message with no level",
		logrus.InfoLevel,
		"message with no level",
	)
	assertLevelAndMessage(
		t,
		logrus.InfoLevel,
		"DEBUG: debug message",
		logrus.DebugLevel,
		"debug message",
	)
	assertLevelAndMessage(
		t,
		logrus.DebugLevel,
		"INFO: info message",
		logrus.InfoLevel,
		"info message",
	)
	assertLevelAndMessage(
		t,
		logrus.DebugLevel,
		"WARN: warning message",
		logrus.WarnLevel,
		"warning message",
	)
	assertLevelAndMessage(
		t,
		logrus.DebugLevel,
		"ERROR: error message",
		logrus.ErrorLevel,
		"error message",
	)
	assertLevelAndMessage(
		t,
		logrus.DebugLevel,
		"17:05:12.779 ERROR: error message with the timestamp",
		logrus.ErrorLevel,
		"error message with the timestamp",
	)
}

func assertLevelAndMessage(
	t *testing.T,
	level logrus.Level,
	message string,
	expectedLevel logrus.Level,
	expectedMessage string,
) {
	actualLevel, actualMessage := getLevelAndMessage(level, message)

	assert.Equal(t, expectedLevel, actualLevel)
	assert.Equal(t, expectedMessage, actualMessage)
}
