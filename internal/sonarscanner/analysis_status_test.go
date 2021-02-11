package sonarscanner

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalysisStatusToString(t *testing.T) {
	assert.Equal(t, "OK", fmt.Sprint(AnalysisStatusOk))
	assert.Equal(t, "WARN", fmt.Sprint(AnalysisStatusWarning))
	assert.Equal(t, "ERROR", fmt.Sprint(AnalysisStatusError))
	assert.Equal(t, "NONE", fmt.Sprint(AnalysisStatusNone))
	assert.Equal(t, "UNDEFINED", fmt.Sprint(AnalysisStatusUndefined))
}

func TestParseAnalysisStatus(t *testing.T) {
	assertAnalysisStatusParsedAs(t, "OK", AnalysisStatusOk)
	assertAnalysisStatusParsedAs(t, "WARN", AnalysisStatusWarning)
	assertAnalysisStatusParsedAs(t, "ERROR", AnalysisStatusError)
	assertAnalysisStatusParsedAs(t, "NONE", AnalysisStatusNone)
}

func TestParseInvalidAnalysisStatus(t *testing.T) {
	status, err := parseAnalysisStatus("UNDEFINED")

	assert.NotNil(t, err)
	assert.Equal(t, AnalysisStatusUndefined, status)
}

func assertAnalysisStatusParsedAs(t *testing.T, value string, expectedStatus AnalysisStatus) {
	actualStatus, err := parseAnalysisStatus(value)

	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, actualStatus)
}
