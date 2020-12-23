package properties

import (
	"bufio"
	"io"
	"os"
	"regexp"
)

// Scanner provides an interface for reading key-value pairs from java properties file.
type Scanner struct {
	scanner *bufio.Scanner
	rx      *regexp.Regexp
	key     string
	value   string
}

// NewScanner creates an instance of PropertiesFileReader.
func NewScanner(reader io.Reader) *Scanner {
	rx, _ := regexp.Compile("^\\s*(([\\w+\\.]+)\\s*=\\s*(.*?))?\\s*(#.*)?$")
	return &Scanner{
		scanner: bufio.NewScanner(reader),
		rx:      rx,
	}
}

// Scan moves to the next property in the properties file.
func (reader *Scanner) Scan() bool {
	for reader.scanner.Scan() {
		matches := reader.rx.FindStringSubmatch(reader.scanner.Text())
		if matches != nil {
			if propertyName := matches[2]; propertyName != "" {
				reader.key = propertyName
				reader.value = matches[3]
				return true
			}
		}
	}

	return false
}

// Err returns the latest file read error.
func (reader *Scanner) Err() error {
	return reader.scanner.Err()
}

// Key returns the last read property name.
func (reader *Scanner) Key() string {
	return reader.key
}

// Value returns the last read property value.
func (reader *Scanner) Value() string {
	return reader.value
}

// ReadProperties makes a property name-value map from properties file stream.
func ReadProperties(reader io.Reader) (map[string]string, error) {
	scanner := NewScanner(reader)
	properties := make(map[string]string)
	for scanner.Scan() {
		properties[scanner.Key()] = scanner.Value()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return properties, nil
}

// ReadPropertiesFile makes a property name-value map from a properties file.
func ReadPropertiesFile(fileName string) (map[string]string, error) {
	if file, err := os.Open(fileName); err != nil {
		return nil, err
	} else {
		defer file.Close()

		return ReadProperties(file)
	}
}
