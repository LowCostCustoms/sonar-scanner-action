package sonarscanner

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/command"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type TaskStatus int

// MetadataFileName is the name of report task file created by the sonar-scanner
// after analysis is finished.
const MetadataFileName = "report-task.txt"

var QualityGateWaitTimeout = errors.New("quality gate wait timeout")

const (
	TaskStatusUndefined  TaskStatus = iota
	TaskStatusPending    TaskStatus = iota
	TaskStatusInProgress TaskStatus = iota
	TaskStatusSuccess    TaskStatus = iota
	TaskStatusFailed     TaskStatus = iota
	TaskStatusCancelled  TaskStatus = iota
)

const (
	taskStatusUndefinedStr  = "UNDEFINED"
	taskStatusPendingStr    = "PENDING"
	taskStatusInProgressStr = "IN_PROGRESS"
	taskStatusSuccessStr    = "SUCCESS"
	taskStatusFailedStr     = "FAILED"
	taskStatusCancelledStr  = "CANCELLED"

	defaultWaitTimeout    = 2 * time.Second
	defaultRequestTimeout = 5 * time.Second

	defaultKeyStoreLocation = "/opt/java/openjdk/lib/security/cacerts"
)

func (status TaskStatus) String() string {
	switch status {
	case TaskStatusPending:
		return taskStatusPendingStr
	case TaskStatusInProgress:
		return taskStatusInProgressStr
	case TaskStatusSuccess:
		return taskStatusSuccessStr
	case TaskStatusFailed:
		return taskStatusFailedStr
	case TaskStatusCancelled:
		return taskStatusCancelledStr
	default:
		return taskStatusUndefinedStr
	}
}

// RunSonarScanner runs the sonar-scanner app.
func RunSonarScanner(
	ctx context.Context,
	entry *logrus.Entry,
	sonarHostURL string,
	scannerWorkingDir string,
	projectFileLocation string,
) error {
	// Override the sonar working directory and metadata file path properties.
	args := []string{
		fmt.Sprintf(
			"-Dsonar.working.directory=%s",
			scannerWorkingDir,
		),
		fmt.Sprintf(
			"-Dsonar.scanner.metadataFilePath=%s",
			path.Join(scannerWorkingDir, MetadataFileName),
		),
	}

	// Override the sonar-server host url if needed.
	if sonarHostURL != "" {
		args = append(
			args,
			fmt.Sprintf("-Dsonar.host.url=%s", sonarHostURL),
		)
	}

	// Set the project file location if needed.
	if projectFileLocation != "" {
		args = append(
			args,
			fmt.Sprintf("-Dproject.settings=%s", projectFileLocation),
		)
	}

	cmd := exec.CommandContext(ctx, "sonar-scanner", args...)

	return command.Run(entry, cmd)
}

// AddCACertificate adds a certificate authority cerificate into the root
// certificate authorities storage.
func AddCACertificate(
	ctx context.Context,
	entry *logrus.Entry,
	certificateData string,
	keyStoreLocation string,
) error {
	if keyStoreLocation == "" {
		keyStoreLocation = defaultKeyStoreLocation
	}

	cmd := exec.CommandContext(
		ctx,
		"keytool",
		"-import",
		"-trustcacerts",
		"-keystore",
		keyStoreLocation,
		"-storepass",
		"changeit",
		"-noprompt",
		"-alias",
		"my-ca-cert",
	)

	cmd.Stdin = strings.NewReader(certificateData)

	return command.Run(entry, cmd)
}

// GetTaskURL reads the value of ceTaskUrl property from the sonar-scanner
// metadata file.
func GetTaskURL(metadataFilePath string) (string, error) {
	file, err := os.Open(metadataFilePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	return getTaskURL(file)
}

func GetTaskStatus(
	ctx context.Context,
	entry *logrus.Entry,
	taskStatusURL string,
) (TaskStatus, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: defaultRequestTimeout,
	}
	for {
		entry.Debugf("reading task status from the server")

		taskStatus, err := requestTaskStatus(ctx, client, taskStatusURL)
		if err != nil {
			return TaskStatusUndefined, err
		}

		entry.Debugf("task status returned by the server %s", taskStatus)

		if taskStatus == TaskStatusSuccess ||
			taskStatus == TaskStatusCancelled ||
			taskStatus == TaskStatusUndefined {
			return taskStatus, nil
		}

		entry.Debugf("waiting for %s before next poll", defaultWaitTimeout)

		select {
		case <-time.After(defaultWaitTimeout):
			continue
		case <-ctx.Done():
			return TaskStatusUndefined, QualityGateWaitTimeout
		}
	}
}

func getTaskURL(reader io.Reader) (string, error) {
	rx, _ := regexp.Compile("^\\s*ceTaskUrl\\s*=\\s*(.*)\\s*$")

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		matches := rx.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return matches[1], nil
		}
	}

	return "", errors.New(
		"could not find the task url in the sonar-scanner metadata file",
	)
}

func requestTaskStatus(
	ctx context.Context,
	client *http.Client,
	taskStatusURL string,
) (TaskStatus, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", taskStatusURL, nil)
	if err != nil {
		return TaskStatusUndefined, err
	}

	response, err := client.Do(request)
	if err != nil {
		if err == context.Canceled {
			return TaskStatusUndefined, QualityGateWaitTimeout
		}

		return TaskStatusUndefined, err
	}

	defer response.Body.Close()

	return processTaskStatusResponse(response)
}

func processTaskStatusResponse(response *http.Response) (TaskStatus, error) {
	if response.StatusCode != http.StatusOK {
		return TaskStatusUndefined, fmt.Errorf(
			"unexpected response code %d",
			response.StatusCode,
		)
	}

	contentType := response.Header.Get("content-type")
	if contentType != "application/json" {
		return TaskStatusUndefined, fmt.Errorf(
			"unexpected response content-type '%s'",
			contentType,
		)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return TaskStatusUndefined, err
	}

	responseJSON := gjson.ParseBytes(responseBody)
	return parseTaskStatus(responseJSON.Get("task.status").Str)
}

func parseTaskStatus(status string) (TaskStatus, error) {
	switch status {
	case taskStatusPendingStr:
		return TaskStatusPending, nil
	case taskStatusInProgressStr:
		return TaskStatusInProgress, nil
	case taskStatusSuccessStr:
		return TaskStatusSuccess, nil
	case taskStatusCancelledStr:
		return TaskStatusCancelled, nil
	case taskStatusFailedStr:
		return TaskStatusFailed, nil
	default:
		return TaskStatusUndefined, fmt.Errorf(
			"unexpected task status '%s'",
			status,
		)
	}
}
