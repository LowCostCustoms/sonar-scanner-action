package main

import (
	"context"
	"path"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/config"
	"github.com/LowCostCustoms/sonar-scanner-action/internal/sonarscanner"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

const (
	defaultSonarWorkingDir = "/opt/sonar-scanner-action/scannerwork/"
)

func init() {
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.DebugLevel
}

func main() {
	env, err := config.GetEnvironment()
	if err != nil {
		log.Fatalf("failed to get the process environment: %s", err)
	}

	importHostCertificate(env)
	runSonarScanner(env)
}

func importHostCertificate(env *config.Environment) {
	if env.SonarHostCertificate != "" {
		log.Info("Importing sonar host certificate ...")

		err := sonarscanner.AddCACertificate(
			context.Background(),
			log.WithField("prefix", "import-certificate"),
			env.SonarHostCertificate,
			"",
		)
		if err != nil {
			log.Fatalf("failed to import sonar host certificate: %s", err)
		}

		log.Info("done")
	} else {
		log.Info("sonar host certificate won't be imported")
	}
}

func runSonarScanner(env *config.Environment) {
	log.Info("running sonar-scanner ...")

	err := sonarscanner.RunSonarScanner(
		context.Background(),
		log.WithField("prefix", "sonar-scanner"),
		env.SonarHostURL,
		defaultSonarWorkingDir,
		env.ProjectFileLocation,
	)
	if err != nil {
		log.Fatalf("failed to run sonar-scanner: %s", err)
	}

	log.Info("done")
}

func waitForQualityGate(env *config.Environment) {
	if env.WaitForQualityGate {
		log.Info("waiting for the quality gate status ...")

		url, err := sonarscanner.GetTaskURL(
			path.Join(
				defaultSonarWorkingDir,
				sonarscanner.MetadataFileName,
			),
		)
		if err != nil {
			log.Fatalf("failed to read sonar-scanner metadata file: %s", err)
		}

		ctx, _ := context.WithTimeout(
			context.Background(),
			env.QualityGateWaitTimeout,
		)
		status, err := sonarscanner.GetTaskStatus(
			ctx,
			log.WithField("prefix", "sonarqube"),
			url,
		)
		if err != nil {
			log.Fatalf("failed to query the analysis results: %s", err)
		}

		if status != sonarscanner.TaskStatusSuccess {
			log.Fatalf("sonarqube task failed with the status %s", status)
		}

		log.Infof("sonarqube task finished with the status %s", status)
	}
}
