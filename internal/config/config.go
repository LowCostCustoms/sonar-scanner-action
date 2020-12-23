package config

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/LowCostCustoms/sonar-scanner-action/internal/properties"
)

type Environment struct {
	SonarHostURL           string
	WaitForQualityGate     bool
	WaitTimeout            time.Duration
	ProjectFileLocation    string
	SonarServerCertificate string
}

type SonarProjectProperties struct {
	SonarHostURL        string
	MetadataFilePath    string
	SonarProjectBaseDir string
}

type CommonConfig struct {
	SonarHostURL     string
	MetadataFilePath string
}

// GetEnvironment returns the process environment variables as a struct.
func GetEnvironment() (*Environment, error) {
	projectFileLocation := os.Getenv("PROJECT_FILE_LOCATION")
	if projectFileLocation == "" {
		projectFileLocation = "sonar-project.properties"
	}

	sonarHostURL := os.Getenv("SONAR_HOST_URL")
	sonarServerCertificate := os.Getenv("SONAR_SERVER_CERTIFICATE")
	waitForQualityGate, err := strconv.ParseBool(os.Getenv("WAIT_FOR_QUALITY_GATE"))
	if err != nil {
		return nil, err
	}

	waitTimeout := time.Duration(0)
	if waitForQualityGate {
		waitTimeout, err = parseDuration(os.Getenv("QUALITY_GATE_WAIT_TIMEOUT"))
		if err != nil {
			return nil, err
		}
	}

	environment := &Environment{
		ProjectFileLocation:    projectFileLocation,
		SonarHostURL:           sonarHostURL,
		SonarServerCertificate: sonarServerCertificate,
		WaitForQualityGate:     waitForQualityGate,
		WaitTimeout:            waitTimeout,
	}
	return environment, nil
}

// ReadSonarProjectProperties reads sonar-scanner project configuration from the specified file and returns it
// as a structure.
func ReadSonarProjectProperties(fileName string) (*SonarProjectProperties, error) {
	props := &SonarProjectProperties{
		MetadataFilePath:    ".scannerwork",
		SonarProjectBaseDir: ".",
	}
	if file, err := os.Open(fileName); err != nil {
		return props, nil
	} else {
		defer file.Close()

		propsMap, err := properties.ReadProperties(file)
		if err != nil {
			return nil, err
		}

		sonarProjectBaseDir := propsMap["sonar.projectBaseDir"]
		if sonarProjectBaseDir != "" {
			props.SonarProjectBaseDir = sonarProjectBaseDir
		}

		metadataFilePath := propsMap["sonar.scanner.metadataFilePath"]
		if metadataFilePath != "" {
			props.MetadataFilePath = metadataFilePath
		} else {
			sonarWorkingDirectory := propsMap["sonar.working.directory"]
			if sonarWorkingDirectory != "" {
				props.MetadataFilePath = sonarWorkingDirectory
			}
		}

		sonarHostURL := propsMap["sonar.host.url"]
		if sonarHostURL != "" {
			props.SonarHostURL = sonarHostURL
		}

		return props, nil
	}
}

// GetCommonConfig returns the sonar-scanner configuration given the environment configuration and sonar-scanner
// project config.
func GetCommonConfig(env *Environment, props *SonarProjectProperties) *CommonConfig {
	config := &CommonConfig{
		SonarHostURL:     props.SonarHostURL,
		MetadataFilePath: path.Join(props.SonarProjectBaseDir, props.MetadataFilePath),
	}
	if env.SonarHostURL != "" {
		config.SonarHostURL = env.SonarHostURL
	}

	return config
}

func parseDuration(value string) (time.Duration, error) {
	rx, _ := regexp.Compile("^(\\d+)(s|m|h)?$")
	matches := rx.FindStringSubmatch(value)
	if matches == nil {
		return time.Duration(0), fmt.Errorf("could not parse duration string '%s'", value)
	}

	duration, _ := strconv.ParseInt(matches[1], 10, 64)
	switch matches[2] {
	case "s":
		return time.Second * time.Duration(duration), nil
	case "m":
		return time.Minute * time.Duration(duration), nil
	case "h":
		return time.Hour * time.Duration(duration), nil
	}

	return time.Duration(0), fmt.Errorf("could not parse duration string '%s'", value)
}
