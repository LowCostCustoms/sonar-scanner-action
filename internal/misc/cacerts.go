package misc

import (
	"context"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

const (
	defaultKeyStoreLocation = "/opt/java/openjdk/lib/security/cacerts"
)

func AddTrustedCertificate(
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

	return RunCommand(entry, cmd)
}
