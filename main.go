package main

import (
	"fmt"
	"log"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/sonarscanner"
	"github.com/LowCostCustoms/sonar-scanner-action/internal/util"
)

func main() {
	env, err := util.GetEnvironment()
	if err != nil {
		log.Fatalf("failed to read the process environment: %s", err)
	}

	projectProperties, err := util.ReadSonarProjectProperties(env.ProjectFileLocation)
	if err != nil {
		log.Fatalf("failed to read sonar-scanner project properties file: %s", err)
	}

	if err := sonarscanner.RunSonarScanner()
}
