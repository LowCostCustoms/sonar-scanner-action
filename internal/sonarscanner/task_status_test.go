package sonarscanner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskStatusToString(t *testing.T) {
	assert.Equal(t, fmt.Sprint(-1), "UNDEFINED")
	assert.Equal(t, fmt.Sprint(TaskStatusSuccess), "SUCCESS")
	assert.Equal(t, fmt.Sprint(TaskStatusPending), "PENDING")
	assert.Equal(t, fmt.Sprint(TaskStatusInProgress), "IN_PROGRESS")
	assert.Equal(t, fmt.Sprint(TaskStatusCancelled), "CANCELLED")
	assert.Equal(t, fmt.Sprint(TaskStatusFailed), "FAILED")
}

func TestParseTaskStatus(t *testing.T) {
	assertTaskStatusParsedAs(t, "PENDING", TaskStatusPending)
	assertTaskStatusParsedAs(t, "IN_PROGRESS", TaskStatusInProgress)
	assertTaskStatusParsedAs(t, "SUCCESS", TaskStatusSuccess)
	assertTaskStatusParsedAs(t, "CANCELLED", TaskStatusCancelled)
	assertTaskStatusParsedAs(t, "FAILED", TaskStatusFailed)
}

func TestParseTaskStatusInvalidInput(t *testing.T) {
	status, err := parseTaskStatus("INVALID")

	assert.NotNil(t, err)
	assert.Equal(t, status, TaskStatusUndefined)

	status, err = parseTaskStatus("smth")

	assert.NotNil(t, err)
	assert.Equal(t, status, TaskStatusUndefined)
}

func assertTaskStatusParsedAs(
	t *testing.T,
	statusString string,
	expectedStatus TaskStatus,
) {
	status, err := parseTaskStatus(statusString)

	assert.Nil(t, err)
	assert.Equal(t, status, expectedStatus)
}
