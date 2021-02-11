package sonarscanner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
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
                    "message": "ok"
                }
                `,
			),
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	json, err := processResponse(response)

	assert.Nil(t, err)
	assert.Equal(t, "ok", json.Get("message").Str)
}

func TestProcessTaskStatusResponseWithInvalidStatus(t *testing.T) {
	response := &http.Response{
		StatusCode: 400,
		Body: &testReader{
			reader: strings.NewReader(
				`
                {
                    "message": "not ok"
                }
                `,
			),
		},
		Header: http.Header{
			"Content-Type": {"application/json"},
		},
	}

	_, err := processResponse(response)

	assert.NotNil(t, err)
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

	_, err := processResponse(response)

	assert.NotNil(t, err)
}

func TestGetTaskUrlFromFile(t *testing.T) {
	tempDir := t.TempDir()
	metadataFileName := path.Join(tempDir, "report-task.txt")

	file, _ := os.Create(metadataFileName)
	file.WriteString("ceTaskUrl=http://poll-me/?q=1")
	file.Close()

	url, err := getTaskUrlFromFile(metadataFileName)

	assert.Nil(t, err)
	assert.Equal(t, url, "http://poll-me/?q=1")
}

func TestGetTaskUrlFromFileInvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	metadataFileName := path.Join(tempDir, "nonexistent-report-task.txt")

	url, err := getTaskUrlFromFile(metadataFileName)

	assert.NotNil(t, err)
	assert.Equal(t, url, "")
}

func TestGetTaskUrl(t *testing.T) {
	reader := strings.NewReader(`
	invalid=invalid
	invalid2=invalid2
	ceTaskUrl=url
	ceTaskUrl_=url2
	`)

	url, err := getTaskUrl(reader)

	assert.Nil(t, err)
	assert.Equal(t, url, "url")
}

func TestGetTaskUrlBadReport(t *testing.T) {
	reader := strings.NewReader("somethingElse=yikes")

	url, err := getTaskUrl(reader)

	assert.NotNil(t, err)
	assert.Equal(t, url, "")
}

func TestNewRun(t *testing.T) {
	factory := RunFactory{
		SonarHostUrl:         "http://localhost",
		SonarHostCert:        "",
		SonarLogin:           "sonar-login",
		SonarPassword:        "sonar-password",
		ProjectFileLocation:  "sonar-project.properties",
		MetadataFileName:     "metadata-file-name",
		ScannerWorkingDir:    "/opt/",
		ScannerVerboseOutput: true,
		LogEntry:             logrus.NewEntry(logrus.New()),
	}

	run, err := factory.NewRun()

	assert.Nil(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, run.sonarHostUrl, "http://localhost")
	assert.Equal(t, run.sonarLogin, "sonar-login")
	assert.Equal(t, run.sonarPassword, "sonar-password")
	assert.Equal(t, run.scannerWorkingDir, "/opt/")
	assert.Equal(t, run.projectFileLocation, "sonar-project.properties")
	assert.Equal(t, run.scannerVerboseOutput, true)
}

func TestNewRunWithProperties(t *testing.T) {
	tempDir := t.TempDir()
	propertiesFileName := path.Join(tempDir, "sonar-project.properties")

	file, _ := os.Create(propertiesFileName)
	file.WriteString(`
	sonar.host.url = http://non-localhost
	sonar.login = login1
	sonar.password = password1
	`)
	file.Close()

	factory := RunFactory{
		ProjectFileLocation: propertiesFileName,
		LogEntry:            logrus.NewEntry(logrus.New()),
	}

	run, err := factory.NewRun()

	assert.Nil(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, run.sonarHostUrl, "http://non-localhost")
	assert.Equal(t, run.sonarLogin, "login1")
	assert.Equal(t, run.sonarPassword, "password1")
}

func TestNewRunPrefersFactorySettings(t *testing.T) {
	tempDir := t.TempDir()
	propertiesFileName := path.Join(tempDir, "sonar-project.properties")

	file, _ := os.Create(propertiesFileName)
	file.WriteString(`
	sonar.host.url = http://non-localhost
	sonar.login = login1
	sonar.password = password1
	`)
	file.Close()

	factory := RunFactory{
		ProjectFileLocation: propertiesFileName,
		SonarHostUrl:        "http://custom-host",
		SonarLogin:          "token",
		LogEntry:            logrus.NewEntry(logrus.New()),
	}

	run, err := factory.NewRun()

	assert.Nil(t, err)
	assert.NotNil(t, run)
	assert.Equal(t, run.sonarHostUrl, "http://custom-host")
	assert.Equal(t, run.sonarLogin, "token")
	assert.Equal(t, run.sonarPassword, "")
}

func TestNewRunWithInvalidSonarHostUrl(t *testing.T) {
	tempDir := t.TempDir()
	propertiesFileName := path.Join(tempDir, "sonar-project.properties")

	file, _ := os.Create(propertiesFileName)
	file.WriteString(`
	sonar.login = login1
	sonar.password = password1
	`)
	file.Close()

	factory := RunFactory{
		ProjectFileLocation: propertiesFileName,
		LogEntry:            logrus.NewEntry(logrus.New()),
	}

	run, err := factory.NewRun()

	assert.Nil(t, run)
	assert.NotNil(t, err)
}

func TestGetSonarScannerArgs(t *testing.T) {
	run := &Run{
		scannerVerboseOutput: true,
		sonarLogin:           "login1",
		sonarPassword:        "password1",
		projectFileLocation:  "props",
		scannerWorkingDir:    "/opt/",
		metadataFilePath:     "mfp",
		sonarHostUrl:         "http://custom-url",
		log:                  logrus.NewEntry(logrus.New()),
	}

	args := run.getSonarScannerArgs()

	assert.Equal(t, len(args), 7)
	assert.Contains(t, args, "-X")
	assert.Contains(t, args, "-Dsonar.login=login1")
	assert.Contains(t, args, "-Dsonar.password=password1")
	assert.Contains(t, args, "-Dproject.settings=props")
	assert.Contains(t, args, "-Dsonar.working.directory=/opt/")
	assert.Contains(t, args, "-Dsonar.scanner.metadataFilePath=/opt/mfp")
	assert.Contains(t, args, "-Dsonar.host.url=http://localhost:6969")
}

func TestGetApiUrl(t *testing.T) {
	assert.Equal(t, "http://host/api/url/", getApiUrl("http://host/", "/api/url/"))
	assert.Equal(t, "http://host/api/url", getApiUrl("http://host", "api/url"))
	assert.Equal(t, "http://host/api/url", getApiUrl("http://host/", "api/url"))
}
