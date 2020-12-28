package sonarscanner

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testReader struct {
	reader io.Reader
}

func (reader *testReader) Read(p []byte) (int, error) {
	return reader.reader.Read(p)
}

func (reader *testReader) Close() error {
	return nil
}

func TestGetTaskUrlSucceeds(t *testing.T) {
	reader := strings.NewReader(
		`
        property=1
        property2=2
        ceTaskUrl  = some url
        `,
	)
	url, err := getTaskUrl(reader)

	assert.Nil(t, err)
	assert.Equal(t, url, "some url")
}

func TestGetTaskUrlFails(t *testing.T) {
	reader := strings.NewReader(
		`
        property=1
        property2=2
        `,
	)
	url, err := getTaskUrl(reader)

	assert.NotNil(t, err)
	assert.Equal(t, url, "")
}

func TestStatusNameFormat(t *testing.T) {
	assert.Equal(t, fmt.Sprint(TaskStatusUndefined), "UNDEFINED")
	assert.Equal(t, fmt.Sprint(TaskStatusPending), "PENDING")
	assert.Equal(t, fmt.Sprint(TaskStatusInProgress), "IN_PROGRESS")
	assert.Equal(t, fmt.Sprint(TaskStatusSuccess), "SUCCESS")
	assert.Equal(t, fmt.Sprint(TaskStatusCancelled), "CANCELLED")
	assert.Equal(t, fmt.Sprint(TaskStatusFailed), "FAILED")
}

func TestProcessTaskStatusResponse(t *testing.T) {
	response := &http.Response{
		StatusCode: 200,
		Body: &testReader{
			reader: strings.NewReader(
				`
                {
                    "task": {
                        "status": "IN_PROGRESS"
                    }
                }
                `,
			),
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	status, err := processTaskStatusResponse(response)

	assert.Nil(t, err)
	assert.Equal(t, status, TaskStatusInProgress)
}

func TestProcessTaskStatusResponseWithInvalidStatus(t *testing.T) {
	response := &http.Response{
		StatusCode: 400,
		Body: &testReader{
			reader: strings.NewReader(
				`
                {
                    "task": {
                        "status": "IN_PROGRESS"
                    }
                }
                `,
			),
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	status, err := processTaskStatusResponse(response)

	assert.NotNil(t, err)
	assert.Equal(t, status, TaskStatusUndefined)
}

func TestProcessTaskStatusResponseWithInvalidContentType(t *testing.T) {
	response := &http.Response{
		StatusCode: 200,
		Body: &testReader{
			reader: strings.NewReader("<html></html>"),
		},
		Header: http.Header{
			"Content-Type": {"text/html"},
		},
	}

	status, err := processTaskStatusResponse(response)

	assert.NotNil(t, err)
	assert.Equal(t, status, TaskStatusUndefined)
}

func TestProcessTaskStatusResponseWithInvalidResponse(t *testing.T) {
	response := &http.Response{
		StatusCode: 200,
		Body: &testReader{
			reader: strings.NewReader(
				`
                {
                    "task": {
                        "noStatus": "error"
                    }
                }
                `,
			),
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	status, err := processTaskStatusResponse(response)

	assert.NotNil(t, err)
	assert.Equal(t, status, TaskStatusUndefined)
}
