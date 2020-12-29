package main

import (
	"context"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/misc"
	"github.com/LowCostCustoms/sonar-scanner-action/internal/sonarscanner"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.InfoLevel
}

func main() {
	env, err := misc.GetEnvironment()
	if err != nil {
		log.Fatalf("Failed to get the process environment: %s", err)
	}

	log.Level = env.LogLevel

	sonarScannerConfig := &sonarscanner.RunConfig{
		TlsSkipVerify:       true,
		SonarHostUrl:        env.SonarHostURL,
		ProjectFileLocation: env.ProjectFileLocation,
		LogEntry:            log.WithField("prefix", "sonar-scanner"),
	}
	sonarScannerRun := sonarScannerConfig.NewRun()

	log.Info("Running sonar-scanner-cli ...")
	err = sonarScannerRun.RunSonarScanner(context.Background())
	if err != nil {
		log.Fatalf("Failed to run sonar-scanner: %s", err)
	}

	if env.WaitForQualityGate {
		log.Info("Retrieving the analysis task status ...")

		ctx, cancel := context.WithTimeout(
			context.Background(),
			env.QualityGateWaitTimeout,
		)
		defer cancel()

		status, err := sonarScannerRun.RetrieveLastAnalysisTaskStatus(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve the task status: %s", err)
		}

		if status != sonarscanner.TaskStatusSuccess {
			log.Fatalf("The analysis task failed with the status %s", status)
		}

		log.Infof("The analysis task finished with the status %s", status)
	}
}
