package properties

import (
	"bufio"
	"io"
	"os"
	"regexp"
)

var rx, _ = regexp.Compile("^\\s*([\\w.]+)\\s*=\\s*(.*?)\\s*(#.*)?$")

type PropertiesMap map[string]string

type PropertiesReader struct {
	scanner *bufio.Scanner
	key     string
	value   string
}

func NewReader(reader io.Reader) *PropertiesReader {
	return &PropertiesReader{
		scanner: bufio.NewScanner(reader),
	}
}

func (reader *PropertiesReader) Scan() bool {
	for reader.scanner.Scan() {
		matches := rx.FindStringSubmatch(reader.scanner.Text())
		if matches != nil {
			reader.key = matches[1]
			reader.value = matches[2]

			return true
		}
	}

	return false
}

func (reader *PropertiesReader) Key() string {
	return reader.key
}

func (reader *PropertiesReader) Value() string {
	return reader.value
}

func (reader *PropertiesReader) Err() error {
	return reader.scanner.Err()
}

func ReadAllProperties(reader io.Reader) (PropertiesMap, error) {
	propertiesReader := NewReader(reader)
	properties := make(PropertiesMap)
	for propertiesReader.Scan() {
		properties[propertiesReader.Key()] = propertiesReader.Value()
	}

	return properties, propertiesReader.Err()
}

func ReadAllPropertiesFromFile(fileName string) (PropertiesMap, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return ReadAllProperties(file)
}
