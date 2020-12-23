package sonarscanner

import (
	"fmt"
	"io"
	"os/exec"
)

// ImportCertificate imports a certficate into the root CA file.
func ImportCertificate(certificateFileStream io.Reader) error {
	command := exec.Command(
		"keytool",
		"-import",
		"-trustedcacerts",
		"-keystore",
		"/opt/java/openjdk/lib/security/cacerts",
		"-storepass",
		"changeit",
		"-noprompt",
		"-alias",
		"custom-ca-cert",
	)
	command.Stdin = certificateFileStream
	return command.Run()
}

// RunSonarScanner runs sonar-scanner analysis and submits results to the server.
func RunSonarScanner(sonarHostURL string) error {
	command := exec.Command("sonar-scanner")
	command.Env = []string{
		fmt.Sprintf("SONAR_HOST_URL=%s", sonarHostURL),
	}
	return command.Run()
}
