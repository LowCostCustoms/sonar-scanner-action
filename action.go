package main

import (
	"context"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/environment"
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
	// Get the process environment.
	env, err := environment.Get()
	if err != nil {
		log.Fatalf("Failed to parse the process environment: %+v", err)
	}

	log.Level = env.LogLevel

	if env.TlsSkipVerify {
		log.Warn("Sonar host certificate verification has beed disabled")
	}

	// Create a new sonar-scanner run.
	runFactory := &sonarscanner.RunFactory{
		SonarHostUrl:        env.SonarHostUrl,
		SonarHostCert:       env.SonarHostCert,
		TlsSkipVerify:       env.TlsSkipVerify,
		ProjectFileLocation: env.ProjectFileLocation,
		SonarLogin:          env.SonarLogin,
		SonarPassword:       env.SonarPassword,
		LogEntry:            log.WithField("prefix", "sonar-scanner"),
	}
	run, err := runFactory.NewRun()
	if err != nil {
		log.Fatalf("Failed to create a sonar scanner run: %s", err)
	}

	// Run sonar-scanner.
	log.Info("Running the sonar scanner cli ...")
	err = run.RunScanner(context.Background())
	if err != nil {
		log.Fatalf("Failed to run sonar scanner: %s", err)
	}

	// Wait for the analysis task result if needed.
	if env.WaitForQualityGate {
		log.Info("Retrieving the analysis task status ...")

		ctx, cancel := context.WithTimeout(context.Background(), env.QualityGateWaitTimeout)
		defer cancel()

		status, err := run.RetrieveLastAnalysisTaskStatus(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve the task status: %s", err)
		}

		if status != sonarscanner.TaskStatusSuccess {
			log.Fatalf("The analysis task failed with the status %s", status)
		}

		log.Infof("The analysis task finished with the status %s", status)
	}
}
