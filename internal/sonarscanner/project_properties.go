package sonarscanner

import (
	"fmt"
	"os"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/properties"
)

type projectProperties struct {
	sonarHostUrl string
	login        string
	password     string
}

func readProjectProperties(projectFile string) (*projectProperties, error) {
	file, err := os.Open(projectFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open properties file: %s", err)
	}

	defer file.Close()

	props := &projectProperties{}
	reader := properties.NewReader(file)
	for reader.Scan() {
		switch reader.Key() {
		case "sonar.host.url":
			props.sonarHostUrl = reader.Value()
			break
		case "sonar.login":
			props.login = reader.Value()
			break
		case "sonar.password":
			props.password = reader.Value()
			break
		}
	}

	return props, reader.Err()
}
