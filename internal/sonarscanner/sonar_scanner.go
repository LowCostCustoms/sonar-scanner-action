package sonarscanner

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
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

	"github.com/LowCostCustoms/sonar-scanner-action/internal/properties"

	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var QualityGateWaitTimeout = errors.New("quality gate wait timeout")

const (
	defaultWaitTimeout         = 2 * time.Second
	defaultRequestTimeout      = 5 * time.Second
	defaultMetadataFileName    = "report-task.txt"
	defaultScannerWorkingDir   = "/opt/sonar-scanner-action/"
	defaultProjectFileLocation = "sonar-project.properties"
	proxyListenAddr            = "localhost:6969"
)

type RunFactory struct {
	SonarHostUrl         string
	SonarHostCert        string
	ScannerWorkingDir    string
	TlsSkipVerify        bool
	MetadataFileName     string
	ProjectFileLocation  string
	SonarLogin           string
	SonarPassword        string
	ScannerVerboseOutput bool
	LogEntry             *logrus.Entry
}

type Run struct {
	sonarHostUrl         string
	scannerWorkingDir    string
	metadataFilePath     string
	projectFileLocation  string
	sonarLogin           string
	sonarPassword        string
	scannerVerboseOutput bool
	tlsConfig            *tls.Config
	log                  *logrus.Entry
}

func (c *RunFactory) NewRun() (*Run, error) {
	metadataFileName := c.MetadataFileName
	if metadataFileName == "" {
		metadataFileName = defaultMetadataFileName
	}

	scannerWorkingDir := c.ScannerWorkingDir
	if scannerWorkingDir == "" {
		scannerWorkingDir = defaultScannerWorkingDir
	}

	projectFileLocation := c.ProjectFileLocation
	if projectFileLocation == "" {
		projectFileLocation = defaultProjectFileLocation
	}

	props, err := c.getProjectProperties()
	if err != nil {
		return nil, err
	}

	tlsConfig, err := c.getTlsClientConfig()
	if err != nil {
		return nil, err
	}

	return &Run{
		sonarHostUrl:         props.sonarHostUrl,
		scannerWorkingDir:    c.ScannerWorkingDir,
		metadataFilePath:     path.Join(scannerWorkingDir, metadataFileName),
		projectFileLocation:  projectFileLocation,
		tlsConfig:            tlsConfig,
		sonarLogin:           props.login,
		sonarPassword:        props.password,
		scannerVerboseOutput: c.ScannerVerboseOutput,
		log:                  c.LogEntry,
	}, nil
}

func (c *RunFactory) getTlsClientConfig() (*tls.Config, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get the system-wide cert pool: %s", err)
	}

	if c.SonarHostCert != "" {
		if !certPool.AppendCertsFromPEM([]byte(c.SonarHostCert)) {
			return nil, fmt.Errorf("failed to append the sonar host certificate to the cert pool")
		}
	}

	return &tls.Config{
		InsecureSkipVerify: c.TlsSkipVerify,
		RootCAs:            certPool,
	}, nil
}

func (c *RunFactory) getProjectProperties() (*projectProperties, error) {
	props := &projectProperties{
		sonarHostUrl: c.SonarHostUrl,
		login:        c.SonarLogin,
		password:     c.SonarPassword,
	}

	if c.ProjectFileLocation != "" {
		if stat, err := os.Stat(c.ProjectFileLocation); err == nil {
			if !stat.IsDir() {
				c.LogEntry.Debugf("Reading project properties from %s", c.ProjectFileLocation)

				projectProps, err := readProjectProperties(c.ProjectFileLocation)
				if err != nil {
					return nil, err
				}

				if props.login == "" {
					c.LogEntry.Debugf("Using sonar scanner auth credentials from the project file")

					props.login = projectProps.login
					props.password = projectProps.password
				}

				if props.sonarHostUrl == "" {
					c.LogEntry.Debugf("Using sonar scanner host from the project file")

					props.sonarHostUrl = projectProps.sonarHostUrl
				}
			} else {
				c.LogEntry.Error("Sonar scanner project file location %s points to a directory", c.ProjectFileLocation)
			}
		} else {
			c.LogEntry.Errorf("Could not open sonar scanner project file %s: %s", c.ProjectFileLocation, err)
		}
	}

	if props.sonarHostUrl == "" {
		return nil, fmt.Errorf("could not infer the sonar host url")
	}

	return props, nil
}

func (c *RunFactory) getSonarHostUrl() (string, error) {
	if c.SonarHostUrl != "" {
		return c.SonarHostUrl, nil
	}

	propertiesMap, err := properties.ReadAllPropertiesFromFile(c.ProjectFileLocation)
	if err != nil {
		return "", err
	}

	sonarHostUrl := propertiesMap["sonar.host.url"]
	if sonarHostUrl == "" {
		return "", fmt.Errorf("could not infer the sonar host url")
	}

	return sonarHostUrl, nil
}

func (r *Run) RunScanner(ctx context.Context) error {
	proxyCtx, proxyCtxCancel := context.WithCancel(ctx)
	defer proxyCtxCancel()

	if err := r.runReverseProxy(proxyCtx); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "sonar-scanner", r.getSonarScannerArgs()...)

	return runSonarScanner(r.log.WithField("prefix", "sonar-scanner-cli"), cmd)
}

func (r *Run) RetrieveLastAnalysisTaskStatus(ctx context.Context) (TaskStatus, error) {
	r.log.Infof("Using metadata file %s", r.metadataFilePath)

	url, err := getTaskUrlFromFile(r.metadataFilePath)
	if err != nil {
		return TaskStatusUndefined, err
	}

	r.log.Infof("Using task result url %s", url)

	return r.retrieveTaskStatus(ctx, url)
}

func (r *Run) runReverseProxy(ctx context.Context) error {
	proxyFactory := &sonarHostProxyFactory{
		listenAddr:   proxyListenAddr,
		config:       r.tlsConfig,
		log:          r.log.WithField("prefix", "sonar-host-proxy"),
		sonarHostUrl: r.sonarHostUrl,
	}
	proxy, err := proxyFactory.new()
	if err != nil {
		return err
	}

	go func() {
		if err := proxy.serveWithContext(ctx); err != nil {
			r.log.Errorf("Failed to start a sonar host proxy: %s", err)
		}
	}()

	return nil
}

func (r *Run) getSonarScannerArgs() []string {
	r.log.Debugf("Sonar-Scanner cli working directory: %s", r.scannerWorkingDir)
	r.log.Debugf("Sonar-Scanner cli metadata file path: %s", r.metadataFilePath)

	args := []string{
		fmt.Sprintf("-Dsonar.working.directory=%s", r.scannerWorkingDir),
		fmt.Sprintf("-Dsonar.scanner.metadataFilePath=%s", path.Join(r.scannerWorkingDir, r.metadataFilePath)),
		fmt.Sprintf("-Dsonar.host.url=http://%s", proxyListenAddr),
	}

	if r.projectFileLocation != "" {
		r.log.Debugf("Sonar-Scanner cli project file location: %s", r.projectFileLocation)

		args = append(args, fmt.Sprintf("-Dproject.settings=%s", r.projectFileLocation))
	}

	if r.sonarLogin != "" {
		args = append(args, fmt.Sprintf("-Dsonar.login=%s", r.sonarLogin))

		if r.sonarPassword != "" {
			args = append(args, fmt.Sprintf("-Dsonar.password=%s", r.sonarPassword))
		}
	}

	if r.scannerVerboseOutput {
		r.log.Debugf("Using sonar-scanner verbose output option")

		args = append(args, "-X")
	}

	return args
}

func (r *Run) retrieveTaskStatus(ctx context.Context, url string) (TaskStatus, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: r.tlsConfig,
		},
		Timeout: defaultRequestTimeout,
	}
	for {
		r.log.Debugf("Reading task status from the server")

		taskStatus, err := r.requestTaskStatus(ctx, client, url)
		if err != nil {
			return TaskStatusUndefined, err
		}

		r.log.Debugf("Task status returned in the response %s", taskStatus)

		if taskStatus == TaskStatusSuccess || taskStatus == TaskStatusCancelled || taskStatus == TaskStatusUndefined {
			return taskStatus, nil
		}

		r.log.Debugf("Waiting for %s before next poll", defaultWaitTimeout)

		select {
		case <-time.After(defaultWaitTimeout):
			continue
		case <-ctx.Done():
			return TaskStatusUndefined, QualityGateWaitTimeout
		}
	}
}

func (r *Run) requestTaskStatus(ctx context.Context, client *http.Client, url string) (TaskStatus, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return TaskStatusUndefined, err
	}

	if r.sonarLogin != "" {
		r.log.Debugf("Using basic auth")

		request.SetBasicAuth(r.sonarLogin, r.sonarPassword)
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

func processTaskStatusResponse(response *http.Response) (TaskStatus, error) {
	if response.StatusCode != http.StatusOK {
		return TaskStatusUndefined, fmt.Errorf("unexpected response code %d", response.StatusCode)
	}

	contentType := response.Header.Get("content-type")
	if contentType != "application/json" {
		return TaskStatusUndefined, fmt.Errorf("unexpected response content-type '%s'", contentType)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return TaskStatusUndefined, err
	}

	responseJSON := gjson.ParseBytes(responseBody)
	return parseTaskStatus(responseJSON.Get("task.status").Str)
}
