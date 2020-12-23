package properties

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FaultyReader struct{}

func (reader *FaultyReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("FaultyReader error")
}

func TestReader(t *testing.T) {
	sourceFile := `
    # this line should be ignored
      property_1 = value_1 # comment
    property_2=value_2#comment
    property.3=#empty
    property.4=
    #property=value #comment
    `

	scanner := NewScanner(strings.NewReader(sourceFile))

	assert.True(t, scanner.Scan())
	assert.Equal(t, scanner.Key(), "property_1")
	assert.Equal(t, scanner.Value(), "value_1")
	assert.True(t, scanner.Scan())
	assert.Equal(t, scanner.Key(), "property_2")
	assert.Equal(t, scanner.Value(), "value_2")
	assert.True(t, scanner.Scan())
	assert.Equal(t, scanner.Key(), "property.3")
	assert.Equal(t, scanner.Value(), "")
	assert.True(t, scanner.Scan())
	assert.Equal(t, scanner.Key(), "property.4")
	assert.Equal(t, scanner.Value(), "")
	assert.False(t, scanner.Scan())
	assert.Nil(t, scanner.Err())
}

func TestReaderPropagatesError(t *testing.T) {
	scanner := NewScanner(&FaultyReader{})

	assert.False(t, scanner.Scan())
	assert.NotNil(t, scanner.Err())
	assert.Equal(t, scanner.Err().Error(), "FaultyReader error")
}

func TestReadProperties(t *testing.T) {
	sourceFile := `
    property.name.first = property_value_first
    property.name.second = property_value_second
    `

	config, err := ReadProperties(strings.NewReader(sourceFile))

	assert.Nil(t, err)
	assert.Equal(t, len(config), 2)
	assert.Equal(t, config["property.name.first"], "property_value_first")
	assert.Equal(t, config["property.name.second"], "property_value_second")
}

func TestReadPropagatesError(t *testing.T) {
	config, err := ReadProperties(&FaultyReader{})

	assert.Nil(t, config)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "FaultyReader error")
}
