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
	"time"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/misc"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var QualityGateWaitTimeout = errors.New("quality gate wait timeout")

const (
	defaultWaitTimeout    = 2 * time.Second
	defaultRequestTimeout = 5 * time.Second
)

type RunConfig struct {
	TlsSkipVerify       bool
	SonarHostUrl        string
	ScannerWorkingDir   string
	MetadataFileName    string
	ProjectFileLocation string
	LogEntry            *logrus.Entry
}

type Run struct {
	tlsSkipVerify       bool
	sonarHostUrl        string
	scannerWorkingDir   string
	metadataFilePath    string
	projectFileLocation string
	log                 *logrus.Entry
}

func (config *RunConfig) CreateRun() *Run {
	metadataFileName := config.MetadataFileName
	if metadataFileName == "" {
		metadataFileName = "report-task.txt"
	}

	scannerWorkingDir := config.ScannerWorkingDir
	if scannerWorkingDir == "" {
		scannerWorkingDir = "/opt/sonar-scanner-action/"
	}

	return &Run{
		tlsSkipVerify:     config.TlsSkipVerify,
		sonarHostUrl:      config.SonarHostUrl,
		scannerWorkingDir: config.ScannerWorkingDir,
		metadataFilePath:  path.Join(scannerWorkingDir, metadataFileName),
		log:               config.LogEntry,
	}
}

func (run *Run) RunSonarScanner(ctx context.Context) error {
	run.log.Debugf(
		"Sonar-Scanner cli working directory: %s",
		run.scannerWorkingDir,
	)
	run.log.Debugf(
		"Sonar-Scanner cli metadata file path: %s",
		run.metadataFilePath,
	)

	args := []string{
		fmt.Sprintf(
			"-Dsonar.working.directory=%s",
			run.scannerWorkingDir,
		),
		fmt.Sprintf(
			"-Dsonar.scanner.metadataFilePath=%s",
			path.Join(run.scannerWorkingDir, run.metadataFilePath),
		),
	}

	if run.sonarHostUrl != "" {
		run.log.Debugf("Sonar-Scanner cli host url: %s", run.sonarHostUrl)

		args = append(
			args,
			fmt.Sprintf("-Dsonar.host.url=%s", run.sonarHostUrl),
		)
	}

	if run.projectFileLocation != "" {
		run.log.Debugf(
			"Sonar-Scanner cli project file location: %s",
			run.projectFileLocation,
		)

		args = append(
			args,
			fmt.Sprintf("-Dproject.settings=%s", run.projectFileLocation),
		)
	}

	cmd := exec.CommandContext(ctx, "sonar-scanner", args...)

	return misc.RunCommand(run.log.WithField("prefix", "sonar-scanner-cli"), cmd)
}

func (run *Run) RetrieveLastAnalysisTaskStatus(
	ctx context.Context,
) (TaskStatus, error) {
	run.log.Infof("Using metadata file %s", run.metadataFilePath)

	url, err := getTaskUrlFromFile(run.metadataFilePath)
	if err != nil {
		return TaskStatusUndefined, err
	}

	run.log.Infof("Using task result url %s", url)

	return run.retrieveTaskStatus(ctx, url)
}

func (run *Run) retrieveTaskStatus(
	ctx context.Context,
	url string,
) (TaskStatus, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: run.tlsSkipVerify,
			},
		},
		Timeout: defaultRequestTimeout,
	}
	for {
		run.log.Debugf("Reading task status from the server")

		taskStatus, err := requestTaskStatus(ctx, client, url)
		if err != nil {
			return TaskStatusUndefined, err
		}

		run.log.Debugf("Task status returned in the response %s", taskStatus)

		if taskStatus == TaskStatusSuccess ||
			taskStatus == TaskStatusCancelled ||
			taskStatus == TaskStatusUndefined {
			return taskStatus, nil
		}

		run.log.Debugf("Waiting for %s before next poll", defaultWaitTimeout)

		select {
		case <-time.After(defaultWaitTimeout):
			continue
		case <-ctx.Done():
			return TaskStatusUndefined, QualityGateWaitTimeout
		}
	}
}

func getTaskUrlFromFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	defer file.Close()

	return getTaskUrl(file)
}

func getTaskUrl(reader io.Reader) (string, error) {
	rx, _ := regexp.Compile("^\\s*ceTaskUrl\\s*=\\s*(.*)\\s*$")

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		matches := rx.FindStringSubmatch(scanner.Text())
		if matches != nil {
			return matches[1], nil
		}
	}

	return "", errors.New("metadata file doesn't contain task url")
}

func requestTaskStatus(
	ctx context.Context,
	client *http.Client,
	url string,
) (TaskStatus, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
