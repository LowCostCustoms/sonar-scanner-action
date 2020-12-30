package properties

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReaderScan(t *testing.T) {
	sourceFile := `
    param_1 = value_1
      param_2=value_2
    param_3  =value_3 # comment
    param_4 = http://some-url/?arg=q
    # comment
    # param_4 = param_5 # comment
    `

	reader := NewReader(strings.NewReader(sourceFile))

	assert.True(t, reader.Scan())
	assert.Equal(t, reader.Key(), "param_1")
	assert.Equal(t, reader.Value(), "value_1")
	assert.True(t, reader.Scan())
	assert.Equal(t, reader.Key(), "param_2")
	assert.Equal(t, reader.Value(), "value_2")
	assert.True(t, reader.Scan())
	assert.Equal(t, reader.Key(), "param_3")
	assert.Equal(t, reader.Value(), "value_3")
	assert.True(t, reader.Scan())
	assert.Equal(t, reader.Key(), "param_4")
	assert.Equal(t, reader.Value(), "http://some-url/?arg=q")
	assert.False(t, reader.Scan())
}
